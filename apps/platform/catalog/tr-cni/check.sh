NS=tr-cni
conf=/etc/cni/net.d/10-kindnet.conflist
plugins_exist() {
  local t
  for t in $(jq -r '.plugins[].type' "$conf"); do [ -x "/opt/cni/bin/$t" ] || return 1; done
}
running() { eq 1 "$(kubectl -n $NS get deployment web -o jsonpath='{.status.readyReplicas}')"; }
has_ip() { [ -n "$(kubectl -n $NS get pods -l app=web -o jsonpath='{.items[0].status.podIP}')" ]; }
untouched() { running && eq 1 "$(kubectl -n $NS get deployment web -o jsonpath='{.metadata.generation}')"; }
plugins_exist && wait_for 120 running

check 2 "Every plugin in the node's CNI config exists in /opt/cni/bin" plugins_exist
check 3 "The Pod of web is Running and Ready" running
check 1 "The Pod of web has an IP address" has_ip
check 1 "Deployment web was not modified" untouched
