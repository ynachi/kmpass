#cloud-config
# vim: syntax=yaml
users:
  - default
  - name: ubuntu
    gecos: Ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin, docker, sudo
    shell: /bin/bash

package_update: true
package_upgrade: true

write_files:
- encoding: b64
  owner: ubuntu:ubuntu
  path: /tmp/install.sh
  permissions: '1551'
  content: {{.NodeBootstrapScript}}
- encoding: b64
  owner: ubuntu:ubuntu
  path: /tmp/haproxy.cfg
  permissions: '0440'
  content: {{.LBConfigFile}}
- encoding: b64
  owner: ubuntu:ubuntu
  path: /tmp/cluster.yaml
  permissions: '1551'
  content: {{.KubeadmInitConfig}}

runcmd:
 - [ sudo, /tmp/install.sh ]
