node=$(kubectl -n nw-pod-facts get pod tracker -o jsonpath='{.spec.nodeName}')
{
  echo "ip: $(kubectl -n nw-pod-facts get pod tracker -o jsonpath='{.status.podIP}')"
  echo "node: $node"
  echo "podCIDR: $(kubectl get node "$node" -o jsonpath='{.spec.podCIDR}')"
} > /opt/course/nw-pod-facts/pod.txt
