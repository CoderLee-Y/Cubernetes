package proxyruntime

import (
	"Cubernetes/pkg/apiserver/crudobj"
	"Cubernetes/pkg/cubeproxy/informer"
	"Cubernetes/pkg/object"
	"errors"
	"github.com/coreos/go-iptables/iptables"
	"log"
	"strconv"
)

const (
	FilterTable  = "filter"
	NatTable     = "nat"
	InputChain   = "INPUT"
	OutputChain  = "OUTPUT"
	DockerChain  = "DOCKER"
	ServiceChain = "SERVICE"
	// SnatOP SNAT use
	SnatOP      = "SNAT"
	PostRouting = "POSTROUTING"

	// DnatOP DNAT use
	DnatOP     = "DNAT"
	PreRouting = "PREROUTING"

	// RANDOM Load balancer policy
	RANDOM      = "random"
	RR          = "nth"
	STATISTIC   = "statistic"
	TestPurpose = false
)

type ProxyRuntime struct {
	Ipt             *iptables.IPTables
	PodInformer     informer.PodInformer
	ServiceInformer informer.ServiceInformer

	ServiceChainMap map[string]ServiceChainElement
}

func (pr *ProxyRuntime) AddService(service *object.Service) error {
	// Check and set default value of service
	err := CheckService(service)
	if err != nil {
		log.Println("Service checking failed, please check service yaml")
		return err
	}

	// init service chain element if NOT EXIST
	if _, ok := pr.ServiceChainMap[service.UID]; ok {
		log.Println("[INFO]: Add existed service, so delete first")
		err = pr.DeleteService(service)
		if err != nil {
			log.Println("Delete service failed")
			return err
		}
	}

	if len(service.Spec.Selector) == 0 {
		log.Println("[INFO]: receive service without selector, cubernetes doesn't support it")
		return nil
	}

	pods, err := crudobj.SelectPods(service.Spec.Selector)
	if err != nil {
		log.Println("Select pods failed")
		return err
	}

	// if pod's ip not filled in, discard it
	for idx, pod := range pods {
		if pod.Status.IP == nil {
			log.Printf("[INFO]: Pod %v can't act as endpoint because no IP allocated", pod.UID)
			pods = append(pods[:idx], pods[idx+1:]...)
		}
	}
	if len(pods) == 0 {
		log.Println("[INFO] No pod to add to service", service.UID)
		return nil
	}

	prob := make([][]string, len(service.Spec.Ports))
	for idx, _ := range prob {
		prob[idx] = make([]string, len(pods))
	}

	pr.ServiceChainMap[service.UID] = ServiceChainElement{
		serviceChainUid:     make([]string, len(service.Spec.Ports)),
		probabilityChainUid: prob,
		numberOfPods:        len(pods),
	}

	for idx, port := range service.Spec.Ports {
		err := pr.MapPortToPods(service, pods, &port, idx)
		if err != nil {
			log.Println("[error]: map port to pods failed")
			return err
		}
	}

	// Write back endpoints
	for _, pod := range pods {
		service.Status.Endpoints = append(service.Status.Endpoints, pod.Status.IP)
	}

	_, err = crudobj.UpdateService(*service)
	if err != nil {
		log.Fatal("[Fatal]: update service failed")
		return err
	}

	return nil
}

// DeleteService It would work even if the service not exist
func (pr *ProxyRuntime) DeleteService(service *object.Service) error {
	if _, ok := pr.ServiceChainMap[service.UID]; !ok {
		log.Println("Delete not exist service")
		return errors.New("delete undef service")
	}

	// delete every
	for idx, port := range service.Spec.Ports {
		err := pr.Ipt.DeleteIfExists(NatTable, ServiceChain,
			"-j", pr.ServiceChainMap[service.UID].serviceChainUid[idx],
			"-d", service.Spec.ClusterIP,
			"-p", string(port.Protocol),
			"--dport", strconv.FormatInt(int64(port.Port), 10))

		if err != nil {
			log.Panicln("Deleting chain failed")
			return err
		}

		err = pr.Ipt.ClearAndDeleteChain(NatTable, pr.ServiceChainMap[service.UID].serviceChainUid[idx])
		if err != nil {
			log.Panicln("Deleting chain failed")
			return err
		}
	}

	for _, servicePort := range pr.ServiceChainMap[service.UID].probabilityChainUid {
		for _, dnat := range servicePort {
			err := pr.Ipt.ClearAndDeleteChain(NatTable, dnat)
			if err != nil {
				return err
			}
		}
	}

	// finally...
	delete(pr.ServiceChainMap, service.UID)
	return nil
}

func (pr *ProxyRuntime) ModifyPod(pod *object.Pod) error {
	services := pr.ServiceInformer.ListServices()

	for _, service := range services {
		if object.MatchLabelSelector(service.Spec.Selector, pod.Labels) {
			err := pr.DeleteService(&service)
			if err != nil {
				return err
			}

			err = pr.AddService(&service)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
