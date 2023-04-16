package main

import (
	"flag"
	"fmt"
	"github.com/ynachi/kmpass/app"
	"os"
)

func main() {
	clusterName := flag.String("cluster", "cluster100", "Name of the kubernetes cluster to deploy.")
	podSubnet := flag.String("pod-subnet", "10.200.0.0/16",
		"Subnet used by pods. Note that this is different from the node's subnet.")
	ctrlNodesNumber := flag.Int("cnodes", 3, "Number of control nodes. Should be minimum 3.")
	workerNodesNumber := flag.Int("wnodes", 1, "Number of worker nodes. Should be minimum 1.")
	ctrlCores := flag.Int("ccores", 2, "Number of control nodes vcpus.")
	workerCores := flag.Int("wcores", 2, "Number of worker nodes vcpus.")
	ctrlMemory := flag.String("cmem", "4G", "Control nodes memory. Support k/K, m/M, g/G suffixes.")
	workerMemory := flag.String("wmem", "4G", "Control nodes memory. Support k/K, m/M, g/G suffixes.")
	lbMemory := flag.String("lmem", "4G", "load-balancer node memory. Support k/K, m/M, g/G suffixes.")
	ctrlDisk := flag.String("cdisk", "20G", "Control nodes disk size. Support k/K, m/M, g/G suffixes."+
		"Its recommended to give at least 20G for a working installation. This field is not yet validated by the tool.")
	workerDisk := flag.String("wdisk", "20G", "Control nodes disk size. Support k/K, m/M, g/G suffixes."+
		"\"Its recommended to give at least 20G for a working installation. This field is not yet validated by the tool.\"")
	lbDisk := flag.String("ldisk", "10G", "Load-balancer node disk size. Support k/K, m/M, g/G suffixes."+
		"\"Its recommended to give at least 10G for a working installation. This field is not yet validated by the tool.\"")
	lbCores := flag.Int("lcores", 2, "Number of lb node vcpus.")
	image := flag.String("image", "20.04", "Ubuntu release version. Only 20.04 works at this time.")
	bootstrapToken := flag.String("token", "5ff0en.1vg4kt1yhk3ty9t7", "Token used to bootstrap the "+
		"cluster. Bootstrap Tokens take the form of abcdef.0123456789abcdef. More formally, they must match the regular "+
		"expression [a-z0-9]{6}\\.[a-z0-9]{16}. They can also be created using the command kubeadm token create.")
	masterKey := "c91d799bfa03fa67107ce07ceb29e67419e5225d4757c93c31ef27bfe8366f0c"
	masterJoinKey := flag.String("ckey", masterKey, "Certificate key used to join the master."+
		"Can be generated using kubeadm certs certificate-key or just use something that matches the format of the"+
		"default key.")
	parallel := flag.Int("parallel", 1, "Number of vms to create concurrently.")
	flag.Parse()

	//@TODO We will use environment variable for the log level
	app.SetLogLevel(app.Debug)

	// 1. Create a cluster configuration
	fmt.Println("------ step 1 ------------")
	cluster := &app.Cluster{
		Name:              *clusterName,
		PodSubnet:         *podSubnet,
		CtrlNodesNumber:   *ctrlNodesNumber,
		CmpNodesNumber:    *workerNodesNumber,
		CtrlNodesMemory:   *ctrlMemory,
		CmpNodesMemory:    *workerMemory,
		CmpNodesCores:     *workerCores,
		CmpNodesDiskSize:  *workerDisk,
		CtrlNodesDiskSize: *ctrlDisk,
		CtrlNodesCores:    *ctrlCores,
		LBNodeMemory:      *lbMemory,
		Image:             *image,
		LBNodeCore:        *lbCores,
		LBNodeDiskSize:    *lbDisk,
		BootstrapToken:    *bootstrapToken,
		KubernetesCertKey: *masterJoinKey,
	}
	if err := cluster.ValidateConfig(); err != nil {
		app.Logger.Error("invalid cluster object", app.ErrClusterConfiguration, "cluster", cluster.Name)
		os.Exit(1)
	}
	// 2. generate cloud init file and get its path
	cloudInitPath, err := app.GenerateConfigCloudInit(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err, "cluster", cluster.Name)
	}
	// 3. create vms, except LB
	cluster.CreateKubeVMs(cloudInitPath, *parallel)
	// 4. generate LB configs
	lbConfPath, err := app.GenerateConfigLB(cluster)
	if err != nil {
		app.Logger.Error("cannot get home dir", err, "cluster", cluster.Name)
		os.Exit(1)
	}
	// 6. Create LB
	_, err = cluster.CreateLB(cloudInitPath, lbConfPath)
	if err != nil {
		app.Logger.Error("cannot create load balancer", err)
		os.Exit(1)
	}
	// 7. generate kubeadm config
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
	// 8. Run kubeadm init on ctrl-0
	err = cluster.KubeInit("/home/ubuntu")
	if err != nil {
		app.Logger.Error("kubeadm init command failed", err, "cluster", cluster.Name)
		os.Exit(1)
	}
	// 9. Install cni
	err = cluster.InstallCNI()
	if err != nil {
		app.Logger.Error("cilium cni installation failed", err, "cluster", cluster.Name)
		os.Exit(1)
	}
	// 10. Join the other controllers
	ctrlJoinCMD, err := cluster.GetMasterJoinCMD()
	if err != nil {
		app.Logger.Error("failed to generate master join command", err, "cluster", cluster.Name)
		os.Exit(1)
	}
	for i := 1; i < cluster.CtrlNodesNumber; i++ {
		_, err = app.RunCmd(fmt.Sprintf("%s-ctrl-%d", cluster.Name, i), ctrlJoinCMD)
		if err != nil {
			app.Logger.Error("unable to join master node", err, "cluster", cluster.Name)
		}
	}
	// 11. Join the workers
	workerJoinCMD, err := cluster.GetWorkerJoinCMD()
	if err != nil {
		app.Logger.Error("failed to generate worker join command", err, "cluster", cluster.Name)
		os.Exit(1)
	}
	for i := 0; i < cluster.CmpNodesNumber; i++ {
		_, err = app.RunCmd(fmt.Sprintf("%s-cmp-%d", cluster.Name, i), workerJoinCMD)
		if err != nil {
			app.Logger.Error("unable to join worker node", err, "cluster", cluster.Name)
		}
	}
}
