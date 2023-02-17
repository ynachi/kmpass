package app

import (
	"os"
	"path/filepath"
	"text/template"
)

// BootstrapConfig will hold k8s cluster independent configuration like nodes bootstrap scripts, LB config files
// and kubernetes version to deploy
type BootstrapConfig struct {
	// Base64 encoded node bootstrap script. Can leverage EncodeFileB64 function for that
	NodeBootstrapScript string
	// Base64 encoded lb config file. Can leverage EncodeFileB64 function for that
	LBConfigFile string
	// @TODO, set kube version
	// KubeVersion string
}

// updateCloudinitNodes updates the nodes cloudinit file. There is a placeholder in this cloudinit file for a
// bootstrap shell script which install the prerequisites for kubernetes. This script can be updated with some
// parameters like Kubernetes version.
func (config BootstrapConfig) updateCloudinitNodesHelper() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger.Error("cannot get home dir", err)
		return ErrGetHomeDirectory
	}
	parsedTemplate, err := template.ParseFiles("app/files/clouds.yaml.tpl")
	if err != nil {
		Logger.Error("unable to load Go template file file", err, "filename", "files/clouds.yaml.tpl")
		return ErrLoadTemplate
	}
	err = os.MkdirAll(homeDir+"/kmpass", 0770)
	if err != nil {
		Logger.Error("unable to load Go template file file", err, "filename", "files/clouds.yaml.tpl")
		return ErrLoadTemplate
	}
	outFilePath := filepath.Join(homeDir, "kmpass", "nodes-cloudinit.yaml")
	outFile, err := os.Create(outFilePath)
	if err != nil {
		Logger.Error("unable to create file", err, "path", outFilePath)
		return ErrCreateFile
	}
	if err := parsedTemplate.Execute(outFile, config); err != nil {
		Logger.Error("unable to parse template", err)
		return ErrParseTemplate
	}
	return nil
}

func UpdateCloudinitNodes(cluster *Cluster) error {
	lbConfPath, err := cluster.generateLBConfig()
	if err != nil {
		Logger.Error("unable to generate lb config file", err, "cluster", cluster.Name)
	}
	bootstrapConfig := BootstrapConfig{}
	encodedBootstrapScript, err := EncodeFileB64("app/files/install.sh")
	if err != nil {
		Logger.Error("unable to encode file", err, "filename", "install.sh")
		return ErrBase64Encode
	}
	encodedLBConfig, err := EncodeFileB64(lbConfPath)
	if err != nil {
		Logger.Error("unable to encode file", err, "filename", "haproxy.cfg")
		return ErrBase64Encode
	}
	bootstrapConfig.NodeBootstrapScript = encodedBootstrapScript
	bootstrapConfig.LBConfigFile = encodedLBConfig
	return bootstrapConfig.updateCloudinitNodesHelper()
}
