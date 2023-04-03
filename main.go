package main

import (
	"fmt"
	"os"

	"github.com/ynachi/kmpass/app"
)

func main() {
	//app.SetLogLevel(app.Error)

	// 1. Create a cluster configuration
	fmt.Println("------ step 1 ------------")
	cluster := &app.Cluster{
		Name:              "cluster100",
		PodSubnet:         "10.200.0.0/16",
		CtrlNodesNumber:   3,
		CmpNodesNumber:    3,
		CtrlNodesMemory:   "4G",
		CmpNodesMemory:    "4G",
		CmpNodesCores:     2,
		CmpNodesDiskSize:  "20G",
		CtrlNodesDiskSize: "20G",
		CtrlNodesCores:    2,
		LBNodeMemory:      "4G",
		Image:             "20.04",
		LBNodeCore:        2,
		LBNodeDiskSize:    "20G",
		BootstrapToken:    "5ff0en.1vg4kt1yhk3ty9t7",
		KubernetesCertKey: "c91d799bfa03fa67107ce07ceb29e67419e5225d4757c93c31ef27bfe8366f0c",
	}
	// 2. generate cloud init file and get it's path
	// @TODO, Check cluster configuration before using it
	fmt.Println("------ step 2 ------------")
	cloudInitPath, err := app.GenerateConfigCloudInit(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
	}
	fmt.Println(cluster)
	fmt.Println("------------------------------")
	// 3. create vms, execpt LB
	fmt.Println("------ step 3 ------------")
	cluster.CreateKubeVMs(cloudInitPath, 2)
	fmt.Println("------------------------------")
	fmt.Println(cluster)
	// 4. generate LB configs
	fmt.Println("------ step 4 ------------")
	lbConfPath, err := app.GenerateConfigLB(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err)
	}
	// 6. Create LB
	fmt.Println("------ step 6 ------------")
	vm, err := cluster.CreateLB(cloudInitPath, lbConfPath)
	fmt.Println(err)
	IP, _ := vm.GetIP()
	fmt.Println(IP)
	fmt.Println(cluster.PublicAPIEndpoint)
	fmt.Println("------------------------------")
	fmt.Println(cluster)
	fmt.Println("Generate and use kubeadm config file")
	// 7. genetrate kubeadm config
	fmt.Println("------ step 7 ------------")
	kubeadmInitConfPath, err := app.GenerateConfigKubeadm(cluster)
	if err != nil {
		os.Exit(1)
	}
	// 7. transfer kubeadm config file to ctrl-0
	firstCtrlNode, err := cluster.GetControlVM(fmt.Sprintf("%s-ctrl-0", cluster.Name), cloudInitPath)
	if err != nil {
		os.Exit(1)
	}
	err = app.Transfer(firstCtrlNode.Name, kubeadmInitConfPath, "cluster.yaml")
	fmt.Println(err)
	// 8. Run kubeadm init on ctrl-0
	fmt.Println("------ step 8 ------------")
	err = cluster.KubeInit("/home/ubuntu")
	if err != nil {
		fmt.Println("Kubeadm init failed")
		os.Exit(1)
	}
	// 9. Install cni
	fmt.Println("------ step 9 ------------")
	err = cluster.InstallCNI()
	if err != nil {
		fmt.Println("unable to install cni")
		os.Exit(1)
	}
	// 10. Join the other controllers
	fmt.Println("------ step 10 ------------")
	ctrlJoinCMD, err := cluster.GetMasterJoinCMD()
	if err != nil {
		fmt.Println("unable to generate control join cmd")
		os.Exit(1)
	}
	for i := 1; i < cluster.CtrlNodesNumber; i++ {
		_, err = app.RunCmd(fmt.Sprintf("%s-ctrl-%d", cluster.Name, i), ctrlJoinCMD)
		if err != nil {
			fmt.Println("unable to join ctrl node")
			os.Exit(1)
		}
	}
	// 11. Join the workers
	fmt.Println("------ step 11 ------------")
	workerJoinCMD, err := cluster.GetWorkerJoinCMD()
	if err != nil {
		fmt.Println("unable to generate worker join cmd")
		os.Exit(1)
	}
	for i := 0; i < cluster.CmpNodesNumber; i++ {
		_, err = app.RunCmd(fmt.Sprintf("%s-ctrl-%d", cluster.Name, i), workerJoinCMD)
		if err != nil {
			fmt.Println("unable to join worker node")
			os.Exit(1)
		}
	}
	hash, _ := cluster.GetCertHash()
	fmt.Println()
	fmt.Println(hash)
	fmt.Println()
	cmd, _ := cluster.GetMasterJoinCMD()
	fmt.Println(cmd)
}
