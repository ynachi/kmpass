package app

import (
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
	if !areValidIPs(cluster.PublicAPIEndpoint) {
		return ErrInvalidIPV4Address
	}
	if !areValidIPs(cluster.CtrlNodesIPs...) {
		return ErrInvalidIPV4Address
	}
	if !areValidIPs(cluster.CmpNodesIPs...) {
		return ErrInvalidIPV4Address
	}
	return nil
}

func areValidIPs(ips ...string) bool {
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
	parsedTpl, err := template.ParseFiles(templatePath)
	filePath := filepath.Join(homeDir, "kmpass", outFileName)
	file, err := os.Create(filePath)
	if err != nil {
		Logger.Error("unable to parse template", err)
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
