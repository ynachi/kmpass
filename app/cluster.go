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
	CmpNodeNumber    int
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
}

// validateConfig checks if cluster configuration is valid.
// It checks Disk sizes, memory sizes, and validity of IP addresses
// @TODO; add more validation (duplicate IPs, PodSubnet, memory and disk sizes, number of nodes)
func (cluster *Cluster) validateConfig() error {
	if !(validateMemoryFormat(cluster.LBNodeMemory) && validateMemoryFormat(cluster.CtrlNodesMemory) && validateMemoryFormat(cluster.CmpNodesMemory)) {
		return ErrMemFormat
	}
	if !(validateMemoryFormat(cluster.LBNodeDiskSize) && validateMemoryFormat(cluster.CtrlNodesDiskSize) && validateMemoryFormat(cluster.CmpNodesDiskSize)) {
		return ErrMemFormat
	}
	if !validateIPs(cluster.PublicAPIEndpoint) {
		return ErrInvalidIPV4Address
	}
	if !validateIPs(cluster.CtrlNodesIPs...) {
		return ErrInvalidIPV4Address
	}
	if !validateIPs(cluster.CmpNodesIPs...) {
		return ErrInvalidIPV4Address
	}
	return nil
}

func validateIPs(ips ...string) bool {
	for _, ip := range ips {
		ipObject := net.ParseIP(ip)
		if ipObject == nil {
			return false
		}
	}
	return true
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
	if err := parsedTpl.Execute(file, *cluster); err != nil {
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
		return lbVM, ErrVMAlreadyExist
	}
	// install lb software packages and transfer lb configuration file
	if _, err := lbVM.RunCmd([]string{"sudo", "apt-get", "install", "haproxy", "-y"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if err := lbVM.Transfer(lbConfPath, "haproxy.cfg"); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if _, err := lbVM.RunCmd([]string{"sudo", "cp", "/tmp/haproxy.cfg", "/etc/haproxy/haproxy.cfg"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	if _, err := lbVM.RunCmd([]string{"sudo", "systemctl", "restart", "haproxy"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return lbVM, err
	}
	Logger.Info("instance created and started with success", "instance-name", lbName)
	IP, err := lbVM.GetIP()
	fmt.Println(IP)
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
		if err := vm.Create(); err != nil {
			Logger.Error("unable to create VM", err, "instance-name", vm.Name)
			return
		}
		IP, err := vm.GetIP()
		if err != nil {
			Logger.Error("unable to retrieve vm IP address", err, "instance-name", vm.Name)
			return
		}
		if strings.Contains(vm.Name, "ctrl") {
			cluster.CtrlNodesIPs = append(cluster.CtrlNodesIPs, IP)
		} else {
			cluster.CmpNodesIPs = append(cluster.CmpNodesIPs, IP)
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
		if !vmCfg.Exist() {
			if err := vmCfg.Create(); err != nil {
				Logger.Error("unable to create control vm instance config", err, "instance-name", vmName)
			}
		} else {
			// @TODO: we should eventually start it but let's keep it this way for now
			Logger.Warn("vm already exist, doing nothing", "instance-name", vmName)
		}
		vms <- vmCfg
	}
	// fill the vms jobs queue with compute nodes
	for i := 0; i < cluster.CmpNodeNumber; i++ {
		vmName := fmt.Sprintf("%s-cmp-%d", cluster.Name, i)
		vmCfg, err := NewInstanceConfig(cluster.CmpNodesCores, cluster.CmpNodesMemory,
			cluster.CmpNodesDiskSize, cluster.Image, vmName, cloudInitPath)
		if err != nil {
			Logger.Error("unable to create compute vm instance config", err, "instance-name", vmName)
		}
		if !vmCfg.Exist() {
			if err := vmCfg.Create(); err != nil {
				Logger.Error("unable to create LB vm instance", err, "instance-name", vmName)
			}
		} else {
			// @TODO: we should eventually start it but let's keep it this way for now
			Logger.Warn("vm already exist, doing nothing", "instance-name", vmName)
		}
		vms <- vmCfg
	}
	close(vms)
	wg.Wait()
}
