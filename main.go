package main

import (
	"github.com/ynachi/kmpass/app"
	"html/template"
	"os"
	"path/filepath"
)

func main() {
	//vm, err := app.New("2", "2G", "20G", "20.04", "yoa-bushit", "/home/ynachi/codes/github.com/kmpass2/app/files/clouds.yaml")
	//if err != nil {
	//	fmt.Println("Failed to create VM")
	//}
	//vm.Create()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
		os.Exit(1)
	}
	encoded, err := app.EncodeFileB64("app/files/install.sh")
	if err != nil {
		app.Logger.Error("unable encode file", err)
	}
	k8sConf := app.Cluster{
		Name:                "cluster100",
		PublicAPIEndpoint:   "172.10.25.2",
		PodSubnet:           "10.200.0.0/16",
		CmpNodesIPs:         []string{"10.10.10.11", "10.10.10.12", "10.10.10.13"},
		CtrlNodesIPs:        []string{"10.10.10.1", "10.10.10.2", "10.10.10.3"},
		NodeBootstrapScript: encoded,
	}
	config, _ := template.ParseFS(app.Files, "files/clouds.yaml.tpl")
	os.MkdirAll(homeDir+"/kmpass", 0770)
	tmpFilePath := filepath.Join(homeDir, ".kmpass", "clouds100.yaml")
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		app.Logger.Error("unable to create temp file", err)
		os.Exit(1)
	}
	if err := config.Execute(tmpFile, k8sConf); err != nil {
		app.Logger.Error("unable to parse template", err)
		os.Exit(1)
	}
}
