---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
networking:
  podSubnet: {{.PodSubnet}}
  dnsDomain: {{.Name}}.local
controlPlaneEndpoint: {{.PublicAPIEndpoint}}:6443
apiServer:
  certSANs:
    - {{.PublicAPIEndpoint}}
clusterName: {{.Name}}
