package object

type Pod struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata"`
	Spec       PodSpec `json:"spec"`
	// use pointer or else omitempty is disabled
	Status *PodStatus `json:"status,omitempty"`
}

type PodSpec struct {
	Containers []Container `json:"containers"`
	Volumes    []Volume    `json:"volumes,omitempty"`
}

type PodStatus struct {
	// reserved for later use
}

type Container struct {
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	// use pointer or else omitempty is disabled
	Resources    *ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts []VolumeMount         `json:"volumeMounts,omitempty"`
	Ports        []ContainerPort       `json:"ports,omitempty"`
}

type ResourceRequirements struct {
	Cpus float64 `json:"cpus,omitempty"`
	// Memory in bytes
	Memory int64 `json:"memory,omitempty"`
}

type Volume struct {
	Name string `json:"name"`
	// Volume only support HostPath type
	HostPath string `json:"hostPath"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

type ContainerPort struct {
	Name          string `json:"name"`
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"hostIP"`
}

type Service struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata"`
	Spec       ServiceSpec   `json:"spec"`
	Status     ServiceStatus `json:"status,omitempty"`
}

type ServiceSpec struct {
	Selector  map[string]string `json:"selector,omitempty"`
	Ports     []ServicePort     `json:"ports,omitempty"`
	ClusterIP string            `json:"ip,omitempty"`
}

type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

type ServicePort struct {
	Protocol   Protocol `json:"protocol,omitempty"`
	Port       int32    `json:"port,omitempty"`
	TargetPort int32    `json:"target,omitempty"`
}

type ServiceStatus struct {
	Ingress []PodIngress `json:"ingress,omitempty"`
}

type PodIngress struct {
	HostName string  `json:"hostname,omitempty"`
	IP       string  `json:"ip,omitempty"`
	Ports    []int32 `json:"ports,omitempty"`
}

type ReplicaSet struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata"`
	Spec       ReplicaSetSpec    `json:"spec"`
	Status     *ReplicaSetStatus `json:"status,omitempty"`
}

type ReplicaSetSpec struct {
	Replicas int32             `json:"replicas"`
	Selector map[string]string `json:"selector,omitempty"`
	Template PodTemplate       `json:"template"`
}

type PodTemplate struct {
	ObjectMeta `json:"metadata"`
	Spec       PodSpec `json:"spec"`
}

type ReplicaSetStatus struct {
	// actual runnig pod replica in PodUIDs
	RunningReplicas int32 `json:"replicas"`
	// UID of pods assigned
	PodUIDs []string `json:"podUIDs"`
}
