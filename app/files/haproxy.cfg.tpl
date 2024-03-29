global
  log /dev/log  local0
  log /dev/log  local1 notice
  chroot /var/lib/haproxy
  stats timeout 30s
  user haproxy
  group haproxy
  daemon

  # Default SSL material locations
  ca-base /etc/ssl/certs
  crt-base /etc/ssl/private

  # Default ciphers to use on SSL-enabled listening sockets.
  # For more information, see ciphers(1SSL). This list is from:
  #  https://hynek.me/articles/hardening-your-web-servers-ssl-ciphers/
  # An alternative list with additional directives can be obtained from
  #  https://mozilla.github.io/server-side-tls/ssl-config-generator/?server=haproxy
  ssl-default-bind-ciphers ECDH+AESGCM:DH+AESGCM:ECDH+AES256:DH+AES256:ECDH+AES128:DH+AES:RSA+AESGCM:RSA+AES:!aNULL:!MD5:!DSS
  ssl-default-bind-options no-sslv3

defaults
  log  global
  mode  tcp
  option  tcplog
  option  dontlognull
        timeout connect 5000
        timeout client  50000
        timeout server  50000
  errorfile 400 /etc/haproxy/errors/400.http
  errorfile 403 /etc/haproxy/errors/403.http
  errorfile 408 /etc/haproxy/errors/408.http
  errorfile 500 /etc/haproxy/errors/500.http
  errorfile 502 /etc/haproxy/errors/502.http
  errorfile 503 /etc/haproxy/errors/503.http
  errorfile 504 /etc/haproxy/errors/504.http


frontend stats
    bind *:80
    mode http
    stats enable
    stats uri /stats
	monitor-uri     /monitoruri
	stats auth admin:change_me_now
    stats refresh 10s

listen kube-api-6443
  bind *:6443
  mode tcp
  {{- range $i, $ip := .CtrlNodesIPs}}
  server ctrl{{$i }} {{$ip -}}:6443 check inter 1s backup
  {{- end}}

listen ingress-router-443
  bind *:443
  mode tcp
  balance source
  {{- range $i, $ip := .CmpNodesIPs}}
  server cmp{{$i }} {{$ip -}}:443 check inter 1s backup
  {{- end}}

listen ingress-router-80
  bind *:80
  mode tcp
  balance source
  {{- range $i, $ip := .CmpNodesIPs}}
  server cmp{{$i }} {{$ip -}}:80 check inter 1s backup
  {{- end}}

