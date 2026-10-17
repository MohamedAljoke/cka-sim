answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/pod.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
pod() { kubectl -n nw-pod-facts get pod tracker -o jsonpath="$1"; }
node() { pod '{.spec.nodeName}'; }

check 1 "ip is the Pod's IP" eq "$(pod '{.status.podIP}')" "$(answer ip)"
check 1 "node is the Pod's node" eq "$(node)" "$(answer node)"
check 1 "podCIDR is that node's Pod CIDR" eq "$(kubectl get node "$(node)" -o jsonpath='{.spec.podCIDR}')" "$(answer podCIDR)"
