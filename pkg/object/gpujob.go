package object

const GpuJobEtcdPrefix = "/apis/gpuJob/"

type GpuJob struct {
	TypeMeta   `json:",inline" yaml:",inline"`
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Status     GpuJobStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type GpuJobStatus struct {
	SlurmJobId string      `json:"slurmJobId,omitempty" yaml:"slurmJobId,omitempty"`
	Phase      GpuJobPhase `json:"phase" yaml:"phase"`
	NodeUID    string
}

type GpuJobPhase string

const (
	JobCreating   GpuJobPhase = "Creating"
	JobCreated    GpuJobPhase = "Created"
	JobSubmitting GpuJobPhase = "Submitting"
	JobWaiting    GpuJobPhase = "Waiting"
	JobRunning    GpuJobPhase = "Running"
	JobSucceeded  GpuJobPhase = "Succeeded"
	JobFailed     GpuJobPhase = "Failed"
)
