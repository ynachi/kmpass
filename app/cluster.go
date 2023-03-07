package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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
	// List of IPs for the compute node. Minimum 1.
	CmpNodesIPs      []string
	CmpNodesMemory   string
	CmpNodesCores    int
	NodesDiskSize    string
	CmpNodesDiskSize string
	// List of IPs for the control node. Minimum 3.
	CtrlNodesIPs      []string
	CtrlNodesMemory   string
	CtrlNodesCores    int
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
// and ready to server traffic.
func (cluster *Cluster) CreateLB(cloudInitPath string, lbConfPath string) error {
	lbName := fmt.Sprintf("%s-lb01", cluster.Name)
	lbVM, err := NewInstanceConfig(cluster.LBNodeCore, cluster.LBNodeMemory, cluster.LBNodeDiskSize, cluster.Image, lbName, cloudInitPath)
	if err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return err
	}
	if !lbVM.Exist() {
		if err := lbVM.Create(); err != nil {
			Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
			return err
		}
	} else {
		// @TODO: we should eventually start it but let's keep it this way for now
		Logger.Warn("vm already exist, doing nothing", "instance-name", lbName)
		return ErrVMAlreadyExist
	}
	// install lb softwares and transfert lb configuration file
	if _, err := lbVM.RunCmd([]string{"sudo", "apt-get", "install", "haproxy", "-y"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return err
	}
	if err := lbVM.Transfer(lbConfPath, "haproxy.cfg"); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return err
	}
	if _, err := lbVM.RunCmd([]string{"sudo", "cp", "/tmp/haproxy.cfg", "/etc/haproxy/haproxy.cfg"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return err
	}
	if _, err := lbVM.RunCmd([]string{"sudo", "systemctl", "restart", "haproxy"}); err != nil {
		Logger.Error("unable to create LB vm instance", err, "instance-name", lbName)
		return err
	}
	Logger.Info("instance created and started with success", "instance-name", lbName)
	return nil
}
