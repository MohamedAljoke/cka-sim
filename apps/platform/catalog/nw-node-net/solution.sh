{
  echo "ip: $(kubectl get node cka-sim-worker2 -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}')"
  echo "gateway: $(ip route show default | awk '{print $3; exit}')"
  echo "ip_forward: $(sysctl -n net.ipv4.ip_forward)"
} > /opt/course/nw-node-net/node.txt
