answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/veth.txt" 2>/dev/null | head -1 | tr -d ' \r' | sed 's/@.*//'; }
pid() { crictl inspect "$(crictl ps -q --name '^probe$' --namespace nw-veth | head -1)" | jq -r '.info.pid'; }
peer() { nsenter -t "$(pid)" -n ip -o link show eth0 | sed -n 's/.*eth0@if\([0-9]*\).*/\1/p'; }
host_if() { ip -o link | awk -F': ' -v i="$(peer)" '$1==i{print $2}' | sed 's/@.*//'; }

check 1 "pid is container probe's main process" eq "$(pid)" "$(answer pid)"
check 1 "pod-ip is the Pod's IP" eq "$(kubectl -n nw-veth get pod probe -o jsonpath='{.status.podIP}')" "$(answer pod-ip)"
check 3 "host-if is the node's end of the Pod's veth pair" eq "$(host_if)" "$(answer host-if)"
