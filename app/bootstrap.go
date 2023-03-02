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
	// @TODO, set kube version
	// KubeVersion string
}

// updateCloudinitNodes updates the nodes cloudinit file. There is a placeholder in this cloudinit file for a
// bootstrap shell script which install the prerequisites for kubernetes. This script can be updated with some
// parameters like Kubernetes version. Returns the cloud init file path in the local machine or an error.
func (config BootstrapConfig) updateCloudinit() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger.Error("cannot get home dir", err)
		return "", ErrGetHomeDirectory
	}
	parsedTemplate, err := template.ParseFiles("app/files/clouds.yaml.tpl")
	if err != nil {
		Logger.Error("unable to load Go template file file", err, "filename", "files/clouds.yaml.tpl")
		return "", ErrLoadTemplate
	}
	err = os.MkdirAll(homeDir+"/kmpass", 0770)
	if err != nil {
		Logger.Error("unable to load Go template file file", err, "filename", "files/clouds.yaml.tpl")
		return "", ErrLoadTemplate
	}
	outFilePath := filepath.Join(homeDir, "kmpass", "cloudinit.yaml")
	outFile, err := os.Create(outFilePath)
	if err != nil {
		Logger.Error("unable to create file", err, "path", outFilePath)
		return "", ErrCreateFile
	}
	if err := parsedTemplate.Execute(outFile, config); err != nil {
		Logger.Error("unable to parse template", err)
		return "", ErrParseTemplate
	}
	return outFilePath, nil
}

// GenerateClusterConfigs generates configuration files used to bootstrap the cluster. These configuration files
// include the lb conf, the kubeadm bootstrap yaml file and the nodes cloudinit.
// returns LBConfPath, kubeadmInitConfPath
func GenerateClusterConfigs(cluster *Cluster) (string, string, string, error) {
	lbConfPath, err := cluster.generateConfigFromTemplate("app/files/haproxy.cfg.tpl", "haproxy.cfg")
	if err != nil {
		Logger.Error("unable to generate lb config file", err, "cluster", cluster.Name)
	}
	kubeadmInitConfPath, err := cluster.generateConfigFromTemplate("app/files/cluster.yaml.tpl", "cluster.yaml")
	if err != nil {
		Logger.Error("unable to generate kubeadm config file", err, "cluster", cluster.Name)
	}
	bootstrapConfig := BootstrapConfig{}
	encodedBootstrapScript, err := EncodeFileB64("app/files/install.sh")
	if err != nil {
		Logger.Error("unable to encode file", err, "filename", "install.sh")
		return "", "", "", ErrBase64Encode
	}
	bootstrapConfig.NodeBootstrapScript = encodedBootstrapScript
	cloudInitPath, err := bootstrapConfig.updateCloudinit()
	return lbConfPath, kubeadmInitConfPath, cloudInitPath, err
}
