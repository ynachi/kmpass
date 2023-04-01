#!/bin/bash

FILE=./k8s_run
if [ -f "$FILE" ]; then
    echo "WARNING!"
    echo "$FILE exists. Script has already been. Do not run on control plane."
    echo "This should be run on the worker node."
    echo
    exit 1
else
    echo "$FILE does not exist. Running  script"
fi


# Create a file when this script is started to keep it from running
# on the control plane node.
sudo touch ./k8s_run

# Update the system
sudo apt update ; sudo apt upgrade -y

# Install necessary software
sudo apt install curl apt-transport-https vim git wget gnupg2 software-properties-common apt-transport-https ca-certificates -y

# Add repo for Kubernetes
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Install the Kubernetes software, and lock the version
sudo apt update
sudo apt -y install kubelet=1.25.5-00 kubeadm=1.25.5-00 kubectl=1.25.5-00
sudo apt-mark hold kubelet kubeadm kubectl

# Ensure Kubelet is running
sudo systemctl enable --now kubelet

# Disable swap just in case
sudo swapoff -a

# Ensure Kernel has modules
sudo modprobe overlay
sudo modprobe br_netfilter

# Update networking to allow traffic
cat <<EOF | sudo tee /etc/sysctl.d/kubernetes.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF

sudo sysctl --system

# Configure containerd settings
cat <<EOF | sudo tee /etc/modules-load.d/containerd.conf
overlay
br_netfilter
EOF

sudo sysctl --system

# Install the containerd software
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt update
sudo apt install containerd.io -y

# Configure containerd and restart
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd

#  Create the config file so no more errors
# Install and configure crictl
export VER="v1.25.0"

wget https://github.com/kubernetes-sigs/cri-tools/releases/download/$VER/crictl-$VER-linux-amd64.tar.gz

tar zxvf crictl-$VER-linux-amd64.tar.gz

sudo mv crictl /usr/local/bin

# Set the endpoints to avoid the deprecation error
sudo crictl config --set \
runtime-endpoint=unix:///run/containerd/containerd.sock \
--set image-endpoint=unix:///run/containerd/containerd.sock

# Add Helm to make our life easier
wget https://get.helm.sh/helm-v3.9.0-linux-amd64.tar.gz
tar -xf helm-v3.9.0-linux-amd64.tar.gz
sudo cp linux-amd64/helm /usr/local/bin/

# Use Cilium as the network plugin
# Install the CLI first
export CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/master/stable.txt)
export CLI_ARCH=amd64

# Ensure correct architecture
if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi
curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

# Make sure download worked
sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum

# Move binary to correct location and remove tarball
sudo tar xzvfC cilium-linux-${CLI_ARCH}.tar.gz /usr/local/bin
rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

# Create ssl certificate if this is ctrl-0 node
if [[ "$(hostname)" == *"-ctrl-0"* ]]; then
  PKI_LOCATION=/etc/kubernetes/pki
  mkdir -p $PKI_LOCATION
  openssl genrsa -out ${PKI_LOCATION}/ca.key 2048
  openssl req -x509 -new -nodes -key ${PKI_LOCATION}/ca.key -subj "/CN=$(hostname)" -days 10000 -out ${PKI_LOCATION}/ca.crt
  openssl x509 -pubkey -in ${PKI_LOCATION}/ca.crt | openssl rsa -pubin -outform der 2>/dev/null \
   | openssl dgst -sha256 -hex | sed 's/^.* /sha256:/' | tee -a /tmp/kubeadm-cert-hash
fi

# ------------ DO THIS ON THE FIRST CTRL NODE ---------------
# sudo kubeadm init --config cluster.yaml --upload-certs
# c91d799bfa03fa67107ce07ceb29e67419e5225d4757c93c31ef27bfe8366f0c
# cilium install

echo '
You can now join any number of the control-plane node running the following command on each as root:

  kubeadm join 10.98.101.94:6443 --token 5ff0en.1vg4kt1yhk3ty9t7 \
        --discovery-token-ca-cert-hash sha256:50630b0116f111ee238f8f4b59e951dccf7f2da52cde8d6bc137542dcb5b63cd \
        --control-plane --certificate-key c91d799bfa03fa67107ce07ceb29e67419e5225d4757c93c31ef27bfe8366f0c


Please note that the certificate-key gives access to cluster sensitive data, keep it secret!
As a safeguard, uploaded-certs will be deleted in two hours; If necessary, you can use
"kubeadm init phase upload-certs --upload-certs" to reload certs afterward.

Then you can join any number of worker nodes by running the following on each as root:

sudo kubeadm join 10.98.101.94:6443 --token 5ff0en.1vg4kt1yhk3ty9t7 \
        --discovery-token-ca-cert-hash sha256:50630b0116f111ee238f8f4b59e951dccf7f2da52cde8d6bc137542dcb5b63cd

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

'

# Ready to continue
sleep 3
echo
echo
echo '***************************'
echo
echo "Continue to the next step"
echo
echo "Use sudo and copy over kubeadm join command from"
echo "control plane."
echo
echo '***************************'
echo
echo
