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
	//app.SetLogLevel(app.Error)

	cluster := &app.Cluster{
		Name:              "cluster100",
		PodSubnet:         "10.200.0.0/16",
		CtrlNodesNumber:   3,
		CmpNodeNumber:     3,
		CtrlNodesMemory:   "4G",
		CmpNodesMemory:    "4G",
		CmpNodesCores:     2,
		CmpNodesDiskSize:  "20G",
		CtrlNodesDiskSize: "20G",
		CtrlNodesCores:    2,
		CmpNodesIPs:       []string{"10.10.10.11", "10.10.10.12", "10.10.10.13"},
		CtrlNodesIPs:      []string{"10.10.10.1", "10.10.10.2", "10.10.10.3"},
		LBNodeMemory:      "4G",
		Image:             "20.04",
		LBNodeCore:        2,
		LBNodeDiskSize:    "20G",
	}
	// @TODO, Check cluster configuration before using it
	cloudInitPath, err := app.GenerateConfigCloudInit(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
	}
	lbConfPath, err := app.GenerateConfigLB(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
	}
	vm, err := cluster.CreateLB(cloudInitPath, lbConfPath)
	fmt.Println(err)
	IP, _ := vm.GetIP()
	fmt.Println(IP)
	fmt.Println(cluster.PublicAPIEndpoint)
	cluster.CreateKubeVMs(cloudInitPath, 1)
}
