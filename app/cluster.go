package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// minCtrlNodes is the minimum number of control nodes
const minCtrlNodes = 3

// minCmpNodes is the minimum number of compute nodes
const minCmpNodes = 1

// Cluster is a struct representing a kubernetes cluster.
// The minimum number of control nodes is 3 and the minimum of compute nodes is 1.
type Cluster struct {
	Name              string
	PublicAPIEndpoint string
	PodSubnet         string
	// List of IPs for the compute node. Minimum 1. They are updated by vm create command. The reason is that at this time,
	// multipass does not support setting VM IP address at creation time. So it has to be retrieved after VM creation time
	// and the cluster info needs to be updated.
	CmpNodesIPs      []string
	CmpNodesMemory   string
	CmpNodesCores    int
	CmpNodesNumber   int
	CmpNodesDiskSize string
	// List of IPs for the control node. Minimum 3.
	CtrlNodesIPs      []string
	CtrlNodesMemory   string
	CtrlNodesCores    int
	CtrlNodesNumber   int
	CtrlNodesDiskSize string
	LBNodeMemory      string
	LBNodeCore        int
	LBNodeDiskSize    string
	// OS image
	Image string
	Mux   sync.Mutex
	// Bootstrap Tokens take the form of abcdef.0123456789abcdef.
	// More formally, they must match the regular expression [a-z0-9]{6}\.[a-z0-9]{16}.
	// They can also be created using the command kubeadm token create.
	BootstrapToken    string
	KubernetesCertKey string
}

// ValidateConfig checks if cluster configuration is valid.
// It checks Disk sizes, memory sizes, and validity of IP addresses
// @TODO; add more validation (duplicate IPs, PodSubnet, memory and disk sizes, number of nodes)
func (cluster *Cluster) ValidateConfig() error {
	// validate memory and disk formats. They use the same function for validation as they share the same format
	if !(validateMemoryFormat(cluster.LBNodeMemory) && validateMemoryFormat(cluster.CtrlNodesMemory) && validateMemoryFormat(cluster.CmpNodesMemory)) {
		Logger.Debug("invalid memory format", "cluster", cluster.Name)
		return ErrMemFormat
	}
	if !(validateMemoryFormat(cluster.LBNodeDiskSize) && validateMemoryFormat(cluster.CtrlNodesDiskSize) && validateMemoryFormat(cluster.CmpNodesDiskSize)) {
		Logger.Debug("invalid disk size format", "cluster", cluster.Name)
		return ErrMemFormat
	}
	// validate control nodes number
	if cluster.CtrlNodesNumber < minCtrlNodes {
		Logger.Debug("a cluster requires at least 3 control nodes", "ctrl-nodes-num", cluster.CtrlNodesNumber)
		return ErrMinControlNodes
	}
	if cluster.CmpNodesNumber%2 == 0 {
		Logger.Debug("an ood number of control nodes is required", "ctrl-nodes-num", cluster.CtrlNodesNumber)
		return ErrOddNumberCtrlNode
	}
	// validate worker nodes number
	if cluster.CmpNodesCores < minCmpNodes {
		Logger.Debug("a cluster requires at least 1 worker nodes", "worker-nodes-num", cluster.CmpNodesNumber)
		return ErrMinComputeNodes
	}
	// validate cluster Pod subnet format
	if _, _, err := net.ParseCIDR(cluster.PodSubnet); err != nil {
		Logger.Debug("cluster Pod subnet address is invalid", "cluster-ip", cluster.PodSubnet)
		return ErrInvalidIPV4Address
	}
	return nil
}

// generateConfigFromTemplate configuration files from templates. Will be used to generate LB and kubernetes
// configurations.
func (cluster *Cluster) generateConfigFromTemplate(templatePath string, outFileName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger.Error("unable to read user home directory", err)
		return "", ErrParseTemplate
	}
	parsedTpl, err := template.ParseFiles(templatePath)
	if err != nil {
		Logger.Error("unable to parse template", err)
		return "", ErrParseTemplate
	}
	filePath := filepath.Join(homeDir, "kmpass", outFileName)
	file, err := os.Create(filePath)
	if err != nil {
		Logger.Error("unable to create file", err, "filename", filePath)
		return filePath, ErrParseTemplate
	}
	if err != nil {
		Logger.Error("unable to create temp file", err)
		return filePath, ErrCreateFile
	}
	if err := parsedTpl.Execute(file, cluster); err != nil {
		Logger.Error("unable to parse template", err)
		return filePath, ErrParseTemplate
	}
	return filePath, nil
}

// CreateLB creates the LB associated with the cluster and run it. After running this method, you'll have a LB deployed
// and ready to server traffic. Returns a pointer to an instance of VM and an error. If the VM already exist, an error
// will be thrown but the VM instance that will be returned will be valid.
func (cluster *Cluster) CreateLB(cloudInitPath string, lbConfPath string) (*Instance, error) {
	lbName := fmt.Sprintf("%s-lb01", cluster.Name)
	lbVM, err := NewInstanceConfig(cluster.LBNodeCore, cluster.LBNodeMemory, cluster.LBNodeDiskSize, cluster.Image, lbName, cloudInitPath)
	if err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if !lbVM.Exist() {
		if err := lbVM.Create(); err != nil {
			Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
			return lbVM, err
		}
	} else {
		// @TODO: we should eventually start it but let's keep it this way for now
		Logger.Warn("vm already exist, doing nothing", "instance-name", lbName)
		IP, err := lbVM.GetIP()
		if err != nil {
			Logger.Error("unable to retrieve vm IP address", err, "instance-name", lbName)
			return lbVM, err
		}
		cluster.PublicAPIEndpoint = IP
		return lbVM, ErrVMAlreadyExist
	}
	// install lb software packages and transfer lb configuration file
	if _, err := RunCmd(lbName, []string{"sudo", "apt-get", "install", "haproxy", "-y"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if err := Transfer(lbName, lbConfPath, "haproxy.cfg"); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if _, err := RunCmd(lbName, []string{"sudo", "cp", "/tmp/haproxy.cfg", "/etc/haproxy/haproxy.cfg"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if _, err := RunCmd(lbName, []string{"sudo", "systemctl", "restart", "haproxy"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	Logger.Info("instance created and started with success", "instance-name", lbName)
	IP, err := lbVM.GetIP()
	if err != nil {
		Logger.Error("unable to retrieve vm IP address", err, "instance-name", lbName)
		return lbVM, err
	}
	cluster.PublicAPIEndpoint = IP
	return lbVM, nil
}

// worker is a helper to create VMs concurrently
func worker(cluster *Cluster, ch <-chan *Instance, wg *sync.WaitGroup) {
	defer wg.Done()
	for vm := range ch {
		if !vm.Exist() {
			if err := vm.Create(); err != nil {
				Logger.Error("unable to create VM", err, "instance-name", vm.Name)
				return
			}
		} else {
			// @TODO: we should eventually start it but let's keep it this way for now
			Logger.Warn("vm already exist, IPs will be updated if needed", "instance-name", vm.Name)
		}
		IP, err := vm.GetIP()
		if err != nil {
			Logger.Error("unable to retrieve vm IP address", err, "instance-name", vm.Name)
			return
		}
		if strings.Contains(vm.Name, "ctrl") {
			cluster.AddControlIP(IP)
		} else {
			cluster.AddComputeIP(IP)
		}
	}
}

// CreateKubeVMs creates a VM for the kubernetes cluster. parallel param is the number of vms to create in parallel.
// This method use the worker concurrency pattern to create and run VMs. Creating a high number of VMs will incur
// network traffic and hypervisor cpu load, so the number of workers should be planned wisely.
func (cluster *Cluster) CreateKubeVMs(cloudInitPath string, numWorkers int) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	vms := make(chan *Instance, numWorkers)

	// create the workers to bootstrap the VMs
	for i := 0; i < numWorkers; i++ {
		go worker(cluster, vms, &wg)
	}

	// fill the vms jobs queue with control nodes
	for i := 0; i < cluster.CtrlNodesNumber; i++ {
		vmName := fmt.Sprintf("%s-ctrl-%d", cluster.Name, i)
		vmCfg, err := NewInstanceConfig(cluster.CtrlNodesCores, cluster.CtrlNodesMemory,
			cluster.CtrlNodesDiskSize, cluster.Image, vmName, cloudInitPath)
		if err != nil {
			Logger.Error("unable to create control vm instance config", err, "instance-name", vmName)
		}
		vms <- vmCfg
	}
	// fill the vms jobs queue with compute nodes
	for i := 0; i < cluster.CmpNodesNumber; i++ {
		vmName := fmt.Sprintf("%s-cmp-%d", cluster.Name, i)
		vmCfg, err := NewInstanceConfig(cluster.CmpNodesCores, cluster.CmpNodesMemory,
			cluster.CmpNodesDiskSize, cluster.Image, vmName, cloudInitPath)
		if err != nil {
			Logger.Error("unable to create compute vm instance config", err, "instance-name", vmName)
		}
		vms <- vmCfg
	}
	close(vms)
	wg.Wait()
}

// containsIP checks if the slice of IP contains the given IP
func containsIP(IPs []string, IP string) bool {
	for _, v := range IPs {
		if v == IP {
			return true
		}
	}
	return false
}

// AddComputeIP add the IP address of a newly created compute machine to compute nodes IP list.
// This is because multipass does not allow to set static IP on a node. So we have to fetch them
// dynamically and update the cluster configurations.
func (cluster *Cluster) AddComputeIP(IP string) {
	if !containsIP(cluster.CmpNodesIPs, IP) {
		cluster.Mux.Lock()
		cluster.CmpNodesIPs = append(cluster.CmpNodesIPs, IP)
		cluster.Mux.Unlock()
	}
}

// AddControlIP add the IP address of a newly created control machine to control nodes IP list.
// This is because multipass does not allow to set static IP on a node. So we have to fetch them
// dynamically and update the cluster configurations.
func (cluster *Cluster) AddControlIP(IP string) {
	if !containsIP(cluster.CtrlNodesIPs, IP) {
		cluster.Mux.Lock()
		cluster.CtrlNodesIPs = append(cluster.CtrlNodesIPs, IP)
		cluster.Mux.Unlock()
	}
}

// GetCertHash retrieves the bootstrap CA certificate hash to populate kubeadm yaml configuration file.
// The hash is generated and stored in the first control node at the path /tmp/kubeadm-cert-hash. The hash
// is directly set as a Cluster attribute. This should be run after the control-0 VM is alive.
func (cluster *Cluster) GetCertHash() (string, error) {
	firstCtrlName := fmt.Sprintf("%s-ctrl-0", cluster.Name)
	return RunCmd(firstCtrlName, []string{"cat", "/tmp/kubeadm-cert-hash"})
}

// InstallCNI installs the cluster CNI. For now, only cilium is supported and is directly hardcoded
// in the tool.
func (cluster *Cluster) InstallCNI() error {
	firstCtrlName := fmt.Sprintf("%s-ctrl-0", cluster.Name)
	_, err := RunCmd(firstCtrlName, []string{"cilium", "install"})
	if err != nil {
		Logger.Error("unable to install cilium cni", err, "cluster", cluster.Name)
	}
	return err
}

// GetMasterJoinCMD build the master join command to join other master nodes
func (cluster *Cluster) GetMasterJoinCMD() ([]string, error) {
	certHash, err := cluster.GetCertHash()
	if err != nil {
		Logger.Error("unable to get cluster certificate hash", err, "cluster-name", cluster.Name)
		return []string{}, err
	}
	cmd := []string{"sudo", "kubeadm", "join"}
	cmd = append(cmd, cluster.PublicAPIEndpoint+":6443")
	cmd = append(cmd, "--token", cluster.BootstrapToken)
	cmd = append(cmd, "--discovery-token-ca-cert-hash", certHash)
	cmd = append(cmd, "--control-plane", "--certificate-key", cluster.KubernetesCertKey)
	return cmd, nil
}

// GetWorkerJoinCMD build the master join command to join other master nodes
func (cluster *Cluster) GetWorkerJoinCMD() ([]string, error) {
	certHash, err := cluster.GetCertHash()
	if err != nil {
		Logger.Error("unable to get cluster certificate hash", err, "cluster-name", cluster.Name)
		return []string{}, err
	}
	cmd := []string{"sudo", "kubeadm", "join"}
	cmd = append(cmd, cluster.PublicAPIEndpoint+":6443")
	cmd = append(cmd, "--token", cluster.BootstrapToken)
	cmd = append(cmd, "--discovery-token-ca-cert-hash", certHash)
	return cmd, nil
}

// GetControlVM return a VM instance populated with a control node parameters.
// This is useful to generate the configuration of an existing vm instance to apply specific instance related
// methods on them (for instance, file transfer)
func (cluster *Cluster) GetControlVM(vmName string, cloudInitPath string) (*Instance, error) {
	vm, err := NewInstanceConfig(
		cluster.CtrlNodesCores,
		cluster.CtrlNodesMemory,
		cluster.CtrlNodesDiskSize,
		cluster.Image,
		vmName,
		cloudInitPath,
	)
	if err != nil {
		Logger.Error("unable to generate control instance config", err, "instance-name", vmName,
			"cluster", cluster.Name)
		return nil, err
	}
	return vm, nil
}

// KubeInit runs kubeadm init cmd from control plane node0
func (cluster *Cluster) KubeInit(remoteHomeDir string) error {
	cmd := []string{"sudo", "kubeadm", "init", "--config", "/tmp/cluster.yaml", "--upload-certs"}
	firstCtrlNodeName := fmt.Sprintf("%s-ctrl-0", cluster.Name)
	if _, err := RunCmd(firstCtrlNodeName, cmd); err != nil {
		Logger.Error("kubeadm init command failed", err, "cluster", cluster.Name)
		return err
	}
	createFolderCmd := []string{"mkdir", "-p", remoteHomeDir + "/.kube"}
	if _, err := RunCmd(firstCtrlNodeName, createFolderCmd); err != nil {
		Logger.Error("cannot create kube config directory", err, "cluster", cluster.Name)
		return err
	}
	copyCmd := []string{"sudo", "cp", "-i", "/etc/kubernetes/admin.conf", remoteHomeDir + "/.kube/config"}
	if _, err := RunCmd(firstCtrlNodeName, copyCmd); err != nil {
		Logger.Error("cannot copy kube auth file to user home directory", err, "cluster", cluster.Name)
		return err
	}
	chownCmd := []string{"sudo", "chown", "1000:1000", "/home/ubuntu/.kube/config"}
	if _, err := RunCmd(firstCtrlNodeName, chownCmd); err != nil {
		Logger.Error("cannot copy kube auth file to user home directory", err, "cluster", cluster.Name)
		return err
	}
	return nil
}
