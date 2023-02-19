---
users:
  - default
  - name: ubuntu
    gecos: Ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin, docker
    shell: /bin/bash
    ssh_import_id: None
    lock_passwd: false
    passwd: $6$rounds=4096$iZ8ODZ/7G9HmUtza$b6v0K347LXlW1tpIUWH47kUM9PUG8wQWNdhpETL446CNHaHloEcLo.1iJreVscrSa8AJiyWD7ZX1/jcpB9dWo.

#cloud-config
package_update: true
package_upgrade: true

write_files:
- encoding: b64
  owner: ubuntu:ubuntu
  path: /tmp/install.sh
  permissions: '1551'
  content: |
    {{.NodeBootstrapScript}}
- encoding: b64
  owner: ubuntu:ubuntu
  path: /tmp/haproxy.cfg
  permissions: '0440'
  content: |
    {{.LBConfigFile}}
- encoding: b64
  owner: ubuntu:ubuntu
  path: /tmp/cluster.yaml
  permissions: '1551'
  content: |
    {{.KubeadmInitConfig}}

runcmd:
 - [ sudo, /tmp/install.sh ]
