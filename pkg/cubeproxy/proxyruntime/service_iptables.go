package proxyruntime

import (
	"Cubernetes/pkg/apiserver/crudobj"
	"Cubernetes/pkg/cubeproxy/utils"
	"Cubernetes/pkg/object"
	"errors"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/google/uuid"
	"log"
	"strconv"
)

/**
@Chenfan
							IPTables
                               XXXXXXXXXXXXXXXXXX
                             XXX     Network    XXX
                               XXXXXXXXXXXXXXXXXX
                                       +
                                       |
                                       v
 +-------------+              +------------------+
 |table: filter| <---+        | table: nat       |
 |chain: INPUT |     |        | chain: PREROUTING|
 +-----+-------+     |        +--------+---------+
       |             |                 |
       v             |                 v
 [local process]     |           ****************          +--------------+
       |             +---------+ Routing decision +------> |table: filter |
       v                         ****************          |chain: FORWARD|
****************                                           +------+-------+
Routing decision                                                  |
****************                                                  |
       |                                                          |
       v                        ****************                  |
+-------------+       +------>  Routing decision  <---------------+
|table: nat   |       |         ****************
|chain: OUTPUT|       |               +
+-----+-------+       |               |
      |               |               v
      v               |      +-------------------+
+--------------+      |      | table: nat        |
|table: filter | +----+      | chain: POSTROUTING|
|chain: OUTPUT |             +--------+----------+
+--------------+                      |
                                      v
                               XXXXXXXXXXXXXXXXXX
                             XXX    Network     XXX
                               XXXXXXXXXXXXXXXXXX
*/

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
	RANDOM    = "random"
	RR        = "nth"
	STATISTIC = "statistic"
)

const TestPurpose = true

type ProxyRuntime struct {
	ipt             *iptables.IPTables
	serviceChainMap map[string]ServiceChainElement
}

func InitIPTables() (*ProxyRuntime, error) {
	pr := &ProxyRuntime{
		ipt:             nil,
		serviceChainMap: make(map[string]ServiceChainElement),
	}

	err := pr.InitObject()
	if err != nil {
		log.Panicln("Init object failed")
		return nil, err
	}

	/* check env */
	flag, err := pr.ipt.ChainExists(FilterTable, DockerChain)
	if !flag {
		log.Printf("Start docker first")
		//return nil, err
	}
	flag, err = pr.ipt.ChainExists(NatTable, DockerChain)
	if !flag {
		log.Printf("Start docker first")
		//return nil, err
	}
	/* Check env ends */

	// Clear all service chain:
	for exist, _ := pr.ipt.Exists(NatTable, PreRouting, "-j", ServiceChain); exist; {
		err := pr.ipt.Delete(NatTable, PreRouting, "-j", ServiceChain)
		if err != nil {
			return nil, err
		}
	}
	for exist, _ := pr.ipt.Exists(NatTable, OutputChain, "-j", ServiceChain); exist; {
		err := pr.ipt.Delete(NatTable, OutputChain, "-j", ServiceChain)
		if err != nil {
			return nil, err
		}
	}

	// Now, create SERVICE CHAIN, and add to PRE-ROUTING/OUTPUT Chain
	// Ref: https://gitee.com/k9-s/Cubernetes/wikis/IPT
	if exists, _ := pr.ipt.ChainExists(NatTable, ServiceChain); !exists {
		err = pr.ipt.NewChain(NatTable, ServiceChain)
		if err != nil {
			log.Panicln("Creating chain failed")
			return nil, err
		}
	}

	err = pr.ipt.Insert(NatTable, PreRouting,
		1, "-j", ServiceChain)
	if err != nil {
		log.Panicln("Creating chain failed")
		return nil, err
	}

	err = pr.ipt.Insert(NatTable, OutputChain, 1,
		"-j", ServiceChain)
	if err != nil {
		log.Panicln("Creating chain failed")
		return nil, err
	}

	return pr, nil
}

// InitObject private function! Just for test
func (pr *ProxyRuntime) InitObject() (err error) {
	pr.ipt, err = iptables.New(iptables.Timeout(3))
	if err != nil {
		log.Println(err)
		return err
	}
	return
}

func (pr *ProxyRuntime) ClearAllService() error {
	for key, _ := range pr.serviceChainMap {
		service, err := crudobj.GetService(key)
		if err != nil {
			return err
		}

		err = pr.DeleteService(&service)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReleaseIPTables Delete all chains in service
func (pr *ProxyRuntime) ReleaseIPTables() error {
	if !TestPurpose {
		err := pr.ClearAllService()
		if err != nil {
			log.Panicln("Error in clear all service.")
			return err
		}
	}

	if exists, _ := pr.ipt.ChainExists(NatTable, ServiceChain); exists {
		err := pr.ipt.Delete(NatTable, OutputChain, "-j", ServiceChain)

		if err != nil {
			log.Println("Error in release IP tables")
			return err
		}

		err = pr.ipt.Delete(NatTable, PreRouting, "-j", ServiceChain)
		if err != nil {
			log.Println("Error in release IP tables")
			return err
		}

		err = pr.ipt.ClearAndDeleteChain(NatTable, ServiceChain)
		if err != nil {
			log.Println("Error in release IP tables")
			return err
		}
	}
	return nil
}

func (pr *ProxyRuntime) AddService(service *object.Service) error {
	// Default value of service
	// any service's cluster IP, modify to pod IP
	err := utils.DefaultService(service)
	if err != nil {
		return err
	}

	pods, err := GetPodByService(service)
	if err != nil || len(pods) == 0 {
		log.Println("Not matched pods found")
		return err
	}

	// init service chain element if NOT EXIST
	if _, ok := pr.serviceChainMap[service.ObjectMeta.UID]; ok {
		err = pr.DeleteService(service)
		if err != nil {
			log.Println("Delete service failed")
			return err
		}
	}

	prob := make([][]string, len(service.Spec.Ports))
	for idx, _ := range prob {
		prob[idx] = make([]string, len(pods))
	}

	pr.serviceChainMap[service.ObjectMeta.UID] = ServiceChainElement{
		serviceChainUid:     make([]string, len(service.Spec.Ports)),
		probabilityChainUid: prob,
		numberOfPods:        len(pods),
	}

	for idx, port := range service.Spec.Ports {
		// Chain name under 29 chars
		serviceUID := fmt.Sprintf("CUBE-SVC-%v", uuid.New().String()[:15])

		// Then create service chain and add to service
		err = pr.ipt.NewChain(NatTable, serviceUID)
		if err != nil {
			log.Println("Create chain failed")
			return err
		}

		// depends on the settings
		// TODO: What if no port/protocol/target port specified?
		err := pr.ipt.Insert(NatTable, ServiceChain,
			1, "-j", serviceUID,
			"-d", service.Spec.ClusterIP,
			"-p", string(port.Protocol),
			"--dport", strconv.FormatInt(int64(port.Port), 10))
		if err != nil {
			log.Panicln("Creating chain failed")
			return err
		}
		pr.serviceChainMap[service.ObjectMeta.UID].serviceChainUid[idx] = serviceUID

		// Then create NUM(pod) chain
		for idx_, pod := range pods {
			podChainUID := fmt.Sprintf("CUBE-SVC-POD-%v", uuid.New().String()[:15])

			err = pr.ipt.NewChain(NatTable, podChainUID)
			if err != nil {
				log.Println("Create chain failed")
				return err
			}

			// if 3 pods, the probability is 0.33/0.50/1.00, so...
			if idx_ < len(pods)-1 {
				probability := float64(1) / float64(len(pods)-idx_)
				err = pr.ipt.Append(NatTable, serviceUID,
					"-j", podChainUID,
					"-m", STATISTIC,
					"--mode", RANDOM,
					"--probability", fmt.Sprintf("%.2f", probability),
				)
				if err != nil {
					log.Println("Create chain failed")
					return err
				}
			} else {
				err = pr.ipt.Append(NatTable, serviceUID,
					"-j", podChainUID,
				)
				if err != nil {
					log.Println("Create chain failed")
					return err
				}
			}

			// at last, add DNAT service
			err = pr.ipt.Insert(NatTable, podChainUID, 1,
				"-j", DnatOP,
				"-p", string(port.Protocol),
				"--to-destination", fmt.Sprintf("%v:%v", pod.Status.IP.String(), strconv.FormatInt(int64(port.TargetPort), 10)),
			)

			if err != nil {
				log.Println("Create chain failed")
				return err
			}
			pr.serviceChainMap[service.ObjectMeta.UID].probabilityChainUid[idx][idx_] = podChainUID

			err = crudobj.AddEndpointToService(service, pod.Status.IP)
			if err != nil {
				log.Println("Update endpoint IP to API Server failed")
				return err
			}
		}
	}

	return nil
}

// DeleteService It would work even if the service not exist
func (pr *ProxyRuntime) DeleteService(service *object.Service) error {
	if _, ok := pr.serviceChainMap[service.ObjectMeta.UID]; !ok {
		log.Println("Delete not exist service")
		return errors.New("delete undef service")
	}

	// delete every
	for idx, port := range service.Spec.Ports {
		err := pr.ipt.DeleteIfExists(NatTable, ServiceChain,
			"-j", pr.serviceChainMap[service.ObjectMeta.UID].serviceChainUid[idx],
			"-d", service.Spec.ClusterIP,
			"-p", string(port.Protocol),
			"--dport", strconv.FormatInt(int64(port.Port), 10))

		if err != nil {
			log.Panicln("Deleting chain failed")
			return err
		}

		err = pr.ipt.ClearAndDeleteChain(NatTable, pr.serviceChainMap[service.ObjectMeta.UID].serviceChainUid[idx])
		if err != nil {
			log.Panicln("Deleting chain failed")
			return err
		}
	}

	for _, servicePort := range pr.serviceChainMap[service.ObjectMeta.UID].probabilityChainUid {
		for _, dnat := range servicePort {
			err := pr.ipt.ClearAndDeleteChain(NatTable, dnat)
			if err != nil {
				return err
			}
		}
	}

	// finally...
	delete(pr.serviceChainMap, service.ObjectMeta.UID)
	return nil
}

func (pr *ProxyRuntime) AddPodAsEndpoints(pod *object.Pod) error {
	services, err := crudobj.GetServices()
	if err != nil {
		log.Println("Get services failed")
		return err
	}

	for _, service := range services {
		if utils.MatchServiceAndPod(&service, pod) {
			// TODO: Finish this function
			err := pr.reshuffleServiceIPTable(&service)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (pr *ProxyRuntime) reshuffleServiceIPTable(service *object.Service) error {
	return nil
}
