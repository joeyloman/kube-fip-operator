package app

type HarvesterConfigStruct struct {
	Metadata    MetadataStruct `json:"metadata"`
	NetworkName string         `json:"networkName"`
}

type HarvesterConfigsStruct struct {
	Items []HarvesterConfigStruct `json:"items"`
}

type MachineConfigRefStruct struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type MachinePoolsStruct struct {
	Name             string                 `json:"name"`
	ControlPlaneRole bool                   `json:"controlPlaneRole"`
	EtcdRole         bool                   `json:"etcdRole"`
	WorkerRole       bool                   `json:"workerRole"`
	MachineConfigRef MachineConfigRefStruct `json:"machineConfigRef"`
}

type RkeConfigStruct struct {
	MachinePools []MachinePoolsStruct `json:"machinePools"`
}

type StatusStruct struct {
	ClusterName string `json:"clusterName"`
}

type SpecStruct struct {
	CloudCredentialSecretName string          `json:"cloudCredentialSecretName"`
	RkeConfig                 RkeConfigStruct `json:"rkeConfig"`
}

type MetadataStruct struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type ClusterStruct struct {
	Metadata MetadataStruct `json:"metadata"`
	Spec     SpecStruct     `json:"spec"`
	Status   StatusStruct   `json:"status"`
}

type ClustersStruct struct {
	ApiVersion string          `json:"apiVersion"`
	Items      []ClusterStruct `json:"items"`
	Kind       string          `json:"kind"`
	Metadata   MetadataStruct  `json:"metadata"`
}

type Cluster struct {
	CloudCredentialSecretName string            `json:"CloudCredentialSecretName"`
	ClusterName               string            `json:"ClusterName"`
	MachineConfigRefName      string            `json:"MachineConfigRefName"`
	Labels                    map[string]string `json:"Labels"`
}
