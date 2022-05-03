package cubelet

import (
	watchobj "Cubernetes/pkg/apiserver/watchobj"
	"Cubernetes/pkg/cubelet/container"
	cuberuntime "Cubernetes/pkg/cubelet/cuberuntime"
	"log"
	"os"
)

type Cubelet struct {
	runtime cuberuntime.CubeRuntime
}

func (cl *Cubelet) syncLoop() {
	ch, cancel, err := watchobj.WatchPods()
	if err != nil {
		log.Panic("Error occurs when watching pods")
		os.Exit(0)
	}

	defer cancel()

	for podEvent := range ch {
		switch podEvent.EType {
		case watchobj.EVENT_PUT:
			err := cl.runtime.SyncPod(&podEvent.Pod, &container.PodStatus{})
			if err != nil {
				return
			}
		case watchobj.EVENT_DELETE:
			err := cl.runtime.KillPod(podEvent.Pod.UID)
			if err != nil {
				return
			}
		default:
			log.Panic("Unsupported type in watch pod.")
		}
	}
}

func (cl *Cubelet) Run() {
	if cl.runtime == nil {
		runtime, err := cuberuntime.NewCubeRuntimeManager()
		if err != nil {
			panic(err)
		}

		cl.runtime = runtime
	}

	defer cl.runtime.Close()
	cl.syncLoop()

	log.Fatalln("Unreachable here")
}
