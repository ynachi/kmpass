package main

import (
	"fmt"
	"github.com/ynachi/kmpass/app"
	"os"
	"path/filepath"
)

func main() {
	//vm, err := app.NewInstanceConfig("2", "2G", "20G", "20.04", "yoa-bushit", "/home/ynachi/codes/github.com/kmpass2/app/files/clouds.yaml")
	//if err != nil {
	//	fmt.Println("Failed to create VM")
	//}
	//vm.Create()
	cluster := &app.Cluster{
		Name:              "cluster100",
		PublicAPIEndpoint: "172.10.25.2",
		PodSubnet:         "10.200.0.0/16",
		CmpNodesIPs:       []string{"10.10.10.11", "10.10.10.12", "10.10.10.13"},
		CtrlNodesIPs:      []string{"10.10.10.1", "10.10.10.2", "10.10.10.3"},
	}
	err := app.UpdateCloudinitNodes(cluster)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
	}
	cloudInitPath := filepath.Join(homeDir, "kmpass", "cloudinit.yaml")
	vmCfg, err := app.NewInstanceConfig("2", "8G", "20G", "20.04", "v500", cloudInitPath)
	fmt.Println(err)
	vmCfg.Create()
}
