package app

// for mapping harvesterconfigs.rke-machine-config.cattle.io
type HarvesterConfigStruct struct {
	Metadata    MetadataStruct `json:"metadata"`
	NetworkInfo string         `json:"networkInfo"`
	NetworkName string         `json:"networkName"`
}

// for mapping harvesterconfigs.rke-machine-config.cattle.io
type HarvesterConfigsStruct struct {
	Items []HarvesterConfigStruct `json:"items"`
}

type HarvesterNetworkInfoInterfacesStruct struct {
	NetworkName string `json:"networkName"`
	MacAddress  string `json:"macAddress"`
}

type HarvesterNetworkInfoStruct struct {
	Interfaces []HarvesterNetworkInfoInterfacesStruct `json:"interfaces"`
}

// for mapping cluster.provisioning.cattle.io
type MachineConfigRefStruct struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// for mapping cluster.provisioning.cattle.io
type MachinePoolsStruct struct {
	Name             string                 `json:"name"`
	ControlPlaneRole bool                   `json:"controlPlaneRole"`
	EtcdRole         bool                   `json:"etcdRole"`
	WorkerRole       bool                   `json:"workerRole"`
	MachineConfigRef MachineConfigRefStruct `json:"machineConfigRef"`
}

// for mapping cluster.provisioning.cattle.io
type RkeConfigStruct struct {
	MachinePools []MachinePoolsStruct `json:"machinePools"`
}

// for mapping cluster.provisioning.cattle.io
type StatusStruct struct {
	ClusterName string `json:"clusterName"`
}

// for mapping cluster.provisioning.cattle.io
type SpecStruct struct {
	CloudCredentialSecretName string          `json:"cloudCredentialSecretName"`
	RkeConfig                 RkeConfigStruct `json:"rkeConfig"`
}

// for mapping cluster.provisioning.cattle.io
type MetadataStruct struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

// for mapping cluster.provisioning.cattle.io
type ClusterStruct struct {
	Metadata MetadataStruct `json:"metadata"`
	Spec     SpecStruct     `json:"spec"`
	Status   StatusStruct   `json:"status"`
}

// for mapping cluster.provisioning.cattle.io
type ClustersStruct struct {
	ApiVersion string          `json:"apiVersion"`
	Items      []ClusterStruct `json:"items"`
	Kind       string          `json:"kind"`
	Metadata   MetadataStruct  `json:"metadata"`
}

// for mapping cluster.management.cattle.io
type SpecManagementStruct struct {
	DisplayName string `json:"displayName"`
}

// for mapping cluster.management.cattle.io
type ClusterManagementStruct struct {
	ApiVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Metadata   MetadataStruct       `json:"metadata"`
	Spec       SpecManagementStruct `json:"spec"`
}

// kube-fip internal
type Cluster struct {
	CloudCredentialSecretName string            `json:"CloudCredentialSecretName"`
	HarvesterClusterName      string            `json:"HarvesterClusterName"`
	ClusterName               string            `json:"ClusterName"`
	MachineConfigRefName      string            `json:"MachineConfigRefName"`
	Labels                    map[string]string `json:"Labels"`
}
