package app

// Cluster is a struct representing a kubernetes cluster.
// We need the IP addresses of the nodes but for now, only the LB one (aka PublicAPIEndpoint) is required. The others
// will be inferred using IPs contiguous to PublicAPIEndpoint.
type Cluster struct {
	Name              string
	PublicAPIEndpoint string
	CmpNodesNumber    int
	CmpNodesMemory    string
	CmpNodesDiskSize  string
	CtrlNodesNumber   int
	CtrlNodesMemory   string
	CtrlNodesCores    int
	CtrlNodesDiskSize string
	LBNodeMemory      string
	LBNodeCore        int
	LBNodeDiskSize    string
	// KubeVersion string
}

// GenerateConfig creates a kubernetes cluster config file to be used by kubeadm init.
// The configuration file is generated via a template and uploaded to the ctrl nodes via cloud-init.
func (cluster *Cluster) GenerateConfig() {

}
