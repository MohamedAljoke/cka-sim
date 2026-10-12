Every node owns one Pod CIDR (`spec.podCIDR`). To reach a Pod on another node, a packet must first
get to that node. A simple CNI like **kindnet** (or flannel's `host-gw` mode) does this with plain
routes, one per other node: "to that node's Pod CIDR, go via that node's IP".

```sh
kubectl get nodes -o custom-columns=NAME:.metadata.name,IP:.status.addresses[0].address,CIDR:.spec.podCIDR
ip route                                     # on the node
#   10.244.0.0/24 via 172.18.0.2 dev eth0    <- control plane's Pods
#   10.244.2.0/24 via 172.18.0.4 dev eth0    <- worker2's Pods
#   10.244.1.2 dev vethXXXX scope host       <- each local Pod: straight into its veth
ip route get 10.244.2.5                      # which route a given address takes
```

This only works when all nodes share one L2 network. Across subnets or clouds, CNIs encapsulate
instead (VXLAN in flannel's default mode, Calico IPIP, Cilium's tunnels) and the route points at a
tunnel device such as `flannel.1` or `cilium_vxlan`. Calico's BGP mode programs routes like these
from a routing daemon.

When Pods on one node can't reach Pods on another, `ip route` on both nodes is one of the first
things to check.
