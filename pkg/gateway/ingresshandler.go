package gateway

import (
	"Cubernetes/pkg/gateway/types"
	"Cubernetes/pkg/object"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (rg *RuntimeGateway) HandleIngress() {
	informEvent := rg.ingressInformer.WatchIngressEvent()

	for ingressEvent := range informEvent {
		log.Printf("[INFO]: Main loop working, types is %v, ingress id is %v",
			ingressEvent.Type, ingressEvent.Ingress.UID)

		switch ingressEvent.Type {
		case types.IngressCreate, types.IngressUpdate:
			log.Printf("[INFO]: Create / Update Ingress, UID is %v", ingressEvent.Ingress.UID)
			// Add to router
			rg.AddIngress(&ingressEvent.Ingress)
		case types.IngressRemove:
			log.Printf("[INFO]: Delete Ingress, UID is %v", ingressEvent.Ingress.UID)
			log.Printf("[Warn]: Not support delete now")
		}
	}
}

func (rg *RuntimeGateway) AddIngress(ingress *object.Ingress) {
	switch ingress.Spec.HTTPType {
	case http.MethodPut:
		rg.router.PUT(ingress.Spec.TriggerPath, rg.GetHandlerByIngress(ingress))
	case http.MethodGet:
		rg.router.GET(ingress.Spec.TriggerPath, rg.GetHandlerByIngress(ingress))
	case http.MethodDelete:
		rg.router.DELETE(ingress.Spec.TriggerPath, rg.GetHandlerByIngress(ingress))
	case http.MethodPost:
		rg.router.POST(ingress.Spec.TriggerPath, rg.GetHandlerByIngress(ingress))
	default:
		log.Printf("[Warn]: unsupported type of http, discard it")
		return
	}

	log.Println("[INFO]: Install ingress successfully")
	if ingress.Status == nil {
		ingress.Status = &object.IngressStatus{}
	}
	ingress.Status.Phase = object.IngressReady
}

// DeleteIngress Test failed: gin not support deleting route on-the-fly
func (rg *RuntimeGateway) DeleteIngress(ingress *object.Ingress) {
	log.Printf("[INFO]: Delete Ingress")
	switch ingress.Spec.HTTPType {
	//case http.MethodPut:
	//	rg.router.PUT(ingress.Spec.TriggerPath, ResourceNotFound)
	//case http.MethodGet:
	//	rg.router.GET(ingress.Spec.TriggerPath, ResourceNotFound)
	//case http.MethodDelete:
	//	rg.router.DELETE(ingress.Spec.TriggerPath, ResourceNotFound)
	//case http.MethodPost:
	//	rg.router.POST(ingress.Spec.TriggerPath, ResourceNotFound)
	default:
		log.Printf("[Warn]: unsupported type of http, discard it")
		return
	}
}

func ResourceNotFound(ctx *gin.Context) {
	ctx.String(http.StatusNotFound, "404 Resource Not Found")
}
