`hostNetwork: true` skips the Pod's own network namespace: no veth and no Pod IP. The containers
share the **node's** interfaces, so the Pod IP equals the node IP and any port it opens is opened
on the node itself.

```sh
kubectl -n nw-hostnetwork get pod node-agent -o wide     # IP = the node's InternalIP
curl http://<worker2-ip>:18080
ss -ltnp | grep 18080                                     # on the node: the process listens there
```

Who uses it: CNI agents (kindnet, calico-node), kube-proxy, and some monitoring agents. They need the
host's network to manage it. Static control-plane Pods use it too, which is why `kube-apiserver`
listens on the node's 6443.

The costs: only one Pod per node can bind a given port (a second one crashes with `address already
in use`, and the scheduler accounts host ports to avoid placing it there), the Pod sees all of the
node's traffic, and NetworkPolicies don't apply to it. Pod Security `baseline` already forbids it.
For a Pod on the default DNS settings with hostNetwork, set `dnsPolicy: ClusterFirstWithHostNet` to
keep cluster DNS.
