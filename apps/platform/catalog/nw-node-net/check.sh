answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/node.txt" 2>/dev/null | head -1 | tr -d ' \r'; }

check 1 "ip is the node's InternalIP" eq "$(kubectl get node "$(hostname)" -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}')" "$(answer ip)"
check 1 "gateway is the default gateway" eq "$(ip route show default | grep -o 'via [^ ]*' | head -1 | cut -d' ' -f2)" "$(answer gateway)"
check 1 "ip_forward is the kernel's value" eq "$(cat /proc/sys/net/ipv4/ip_forward)" "$(answer ip_forward)"
