node=$(hostname)
configz() { kubectl get --raw "/api/v1/nodes/$node/proxy/configz" 2>/dev/null | jq -r ".kubeletconfig.$1"; }
ready() { eq True "$(kubectl get node "$node" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
in_file() { grep -qE '^containerLogMaxSize: *"?5Mi"?$' /var/lib/kubelet/config.yaml && grep -qE '^containerLogMaxFiles: *3$' /var/lib/kubelet/config.yaml; }
running() { ready && eq "5Mi 3" "$(configz containerLogMaxSize) $(configz containerLogMaxFiles)"; }
in_file && wait_for 90 running

check 2 "The kubelet config file sets containerLogMaxSize 5Mi and containerLogMaxFiles 3" in_file
check 3 "The running kubelet uses them and the node is Ready" running
