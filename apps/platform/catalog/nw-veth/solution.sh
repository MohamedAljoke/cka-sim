pid=$(crictl inspect "$(crictl ps -q --name '^probe$' --namespace nw-veth | head -1)" | jq -r .info.pid)
# nsenter -t $pid -n ip link show eth0 -> "2: eth0@if7: ..." : the peer is interface number 7 on the host
peer=$(nsenter -t "$pid" -n ip -o link show eth0 | sed -n 's/.*eth0@if\([0-9]*\).*/\1/p')
{
  echo "pid: $pid"
  echo "pod-ip: $(nsenter -t "$pid" -n ip -4 -o addr show eth0 | awk '{print $4}' | cut -d/ -f1)"
  echo "host-if: $(ip -o link | awk -F': ' -v i="$peer" '$1==i{print $2}' | sed 's/@.*//')"
} > /opt/course/nw-veth/veth.txt
