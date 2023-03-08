package main

import (
	"fmt"

	"github.com/ynachi/kmpass/app"
)

func main() {
	//vm, err := app.NewInstanceConfig("2", "2G", "20G", "20.04", "yoa-bushit", "/home/ynachi/codes/github.com/kmpass2/app/files/clouds.yaml")
	//if err != nil {
	//	fmt.Println("Failed to create VM")
	//}
	//vm.Create()
	app.SetLogLevel(app.Error)
	cluster := &app.Cluster{
		Name:              "cluster100",
		PublicAPIEndpoint: "172.10.25.2",
		PodSubnet:         "10.200.0.0/16",
		CmpNodesIPs:       []string{"10.10.10.11", "10.10.10.12", "10.10.10.13"},
		CtrlNodesIPs:      []string{"10.10.10.1", "10.10.10.2", "10.10.10.3"},
		LBNodeMemory:      "4G",
		Image:             "20.04",
		LBNodeCore:        2,
		LBNodeDiskSize:    "20G",
	}
	lbConfPath, _, cloudInitPath, err := app.GenerateClusterConfigs(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
	}
	err = cluster.CreateLB(cloudInitPath, lbConfPath)
	fmt.Println(err)
}
