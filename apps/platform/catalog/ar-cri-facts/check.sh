answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/runtime.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
runtime_ok() { eq containerd "$(answer runtime | tr 'A-Z' 'a-z')"; }
socket_ok() { [[ "$(answer socket)" =~ ^(unix://)?(/var)?/run/containerd/containerd\.sock$ ]]; }
id_ok() {
  local id
  id=$(answer kube-proxy)
  [ "${#id}" -ge 12 ] && crictl ps -q --name '^kube-proxy$' | grep -q "^$id"
}

check 1 "runtime is containerd" runtime_ok
check 1 "socket is containerd's CRI socket" socket_ok
check 1 "kube-proxy is the ID of the running kube-proxy container" id_ok
