package types

import "Cubernetes/pkg/object"

type PodEventType string

const (
	PodCreate PodEventType = "create"
	PodUpdate PodEventType = "update"
	PodKilled PodEventType = "killed"
)

type PodEvent struct {
	Type PodEventType
	Pod  object.Pod
}

type RsEventType string

const (
	RsCreate RsEventType = "create"
	RsUpdate RsEventType = "update"
	RsRemove RsEventType = "remove"
)

type RsEvent struct {
	Type       RsEventType
	ReplicaSet object.ReplicaSet
}

type AsEventType string

const (
	AsCreate AsEventType = "create"
	AsUpdate AsEventType = "update"
	AsRemove AsEventType = "remove"
)

type AsEvent struct {
	Type       AsEventType
	AutoScaler object.AutoScaler
}
