package cubeconfig

import (
	"time"
)

var APIServerIp = "127.0.0.1"

const ETCDTimeout = time.Second * 2
const ETCDAddr = "127.0.0.1:2379"

const ServiceClusterIPRange = "172.16.0.0/16"

const APIServerPort = 8080
const HeartbeatPort = 8081
const DefaultApiVersion = "v1"
const CubeVersion = "v1.0"

const (
	JobFileDir    = "/etc/cubernetes/apiserver/jobs/"
	ActionFileDir = "/etc/cubernetes/apiserver/actions/"
	StaticDir     = "/etc/cubernetes/static"
	MetaDir       = "/etc/cubernetes/cubernetes/"
	MetaFile      = MetaDir + "meta"
)
