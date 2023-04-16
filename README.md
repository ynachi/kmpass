# kmpass
Kubernetes deployment tool with multipass and Go

Deploy a real multi-node cluster in your developer workstation withoust friction. The deployment 
includes at least 3 control plane nodes and 1 worker node. Every node is built independently (ie no
 golden OS). The build deployment may be slow on a single CPU. That why you can span it across multiple
 CPUs if available in your environment. Cilium CNI is enforced for now.

## Overview of the cli options
```bash
./kmpass-0.1.0-linux-amd64 --help
Usage of ./kmpass-0.1.0-linux-amd64:
  -ccores int
        Number of control nodes vcpus. (default 2)
  -cdisk string
        Control nodes disk size. Support k/K, m/M, g/G suffixes.Its recommended to give at least 20G for a working installation. This field is not yet validated by the tool. (default "20G")
  -ckey string
        Certificate key used to join the master.Can be generated using kubeadm certs certificate-key or just use something that matches the format of thedefault key. (default "c91d799bfa03fa67107ce07ceb29e67419e5225d4757c93c31ef27bfe8366f0c")
  -cluster string
        Name of the kubernetes cluster to deploy. (default "cluster100")
  -cmem string
        Control nodes memory. Support k/K, m/M, g/G suffixes. (default "4G")
  -cnodes int
        Number of control nodes. Should be minimum 3. (default 3)
  -image string
        Ubuntu release version. Only 20.04 works at this time. (default "20.04")
  -lcores int
        Number of lb node vcpus. (default 2)
  -ldisk string
        Load-balancer node disk size. Support k/K, m/M, g/G suffixes."Its recommended to give at least 10G for a working installation. This field is not yet validated by the tool." (default "10G")
  -lmem string
        load-balancer node memory. Support k/K, m/M, g/G suffixes. (default "4G")
  -parallel int
        Number of vms to create concurrently. (default 1)
  -pod-subnet string
        Subnet used by pods. Note that this is different from the node's subnet. (default "10.200.0.0/16")
  -token string
        Token used to bootstrap the cluster. Bootstrap Tokens take the form of abcdef.0123456789abcdef. More formally, they must match the regular expression [a-z0-9]{6}\.[a-z0-9]{16}. They can also be created using the command kubeadm token create. (default "5ff0en.1vg4kt1yhk3ty9t7")
  -wcores int
        Number of worker nodes vcpus. (default 2)
  -wdisk string
        Control nodes disk size. Support k/K, m/M, g/G suffixes."Its recommended to give at least 20G for a working installation. This field is not yet validated by the tool." (default "20G")
  -wmem string
        Control nodes memory. Support k/K, m/M, g/G suffixes. (default "4G")
  -wnodes int
        Number of worker nodes. Should be minimum 1. (default 1)
```

## How to use kmpass
Download a release version for your OS, set it as executable (aka chmod +x) and you are good to go.
I recommend to rename the release `kmpass` and add it to your user bin for easy access.

Example
```bash
# create a cluster named app300 with 3 control, 3 worker and 1 lb nodes and use most of the 
# parameters default values

kmpass --cluster app300 -wnodes 3 --parallel 3

# after it's done, connect to one of the ctrl node and check your k8s installation

multipass exec app300-ctrl-0 bash

ubuntu@app300-ctrl-0:~$ kubectl get nodes -o wide
NAME            STATUS   ROLES           AGE   VERSION   INTERNAL-IP     EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
app300-cmp-0    Ready    <none>          67m   v1.25.5   10.175.83.148   <none>        Ubuntu 20.04.6 LTS   5.4.0-139-generic   containerd://1.6.20
app300-cmp-1    Ready    <none>          67m   v1.25.5   10.175.83.236   <none>        Ubuntu 20.04.6 LTS   5.4.0-139-generic   containerd://1.6.20
app300-cmp-2    Ready    <none>          67m   v1.25.5   10.175.83.137   <none>        Ubuntu 20.04.6 LTS   5.4.0-139-generic   containerd://1.6.20
app300-ctrl-0   Ready    control-plane   70m   v1.25.5   10.175.83.54    <none>        Ubuntu 20.04.6 LTS   5.4.0-139-generic   containerd://1.6.20
app300-ctrl-1   Ready    control-plane   69m   v1.25.5   10.175.83.71    <none>        Ubuntu 20.04.6 LTS   5.4.0-139-generic   containerd://1.6.20
app300-ctrl-2   Ready    control-plane   68m   v1.25.5   10.175.83.44    <none>        Ubuntu 20.04.6 LTS   5.4.0-139-generic   containerd://1.6.20

ubuntu@app300-ctrl-0:~$ cilium status
    /¯¯\
 /¯¯\__/¯¯\    Cilium:          OK
 \__/¯¯\__/    Operator:        OK
 /¯¯\__/¯¯\    Hubble Relay:    disabled
 \__/¯¯\__/    ClusterMesh:     disabled
    \__/

DaemonSet         cilium             Desired: 6, Ready: 6/6, Available: 6/6
Deployment        cilium-operator    Desired: 1, Ready: 1/1, Available: 1/1
Containers:       cilium             Running: 6
                  cilium-operator    Running: 1
Cluster Pods:     2/2 managed by Cilium
Image versions    cilium             quay.io/cilium/cilium:v1.13.1@sha256:428a09552707cc90228b7ff48c6e7a33dc0a97fe1dd93311ca672834be25beda: 6
                  cilium-operator    quay.io/cilium/operator-generic:v1.13.1@sha256:f47ba86042e11b11b1a1e3c8c34768a171c6d8316a3856253f4ad4a92615d555: 1
```