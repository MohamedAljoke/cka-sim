Three facts every node needs right before Kubernetes networking can work:

```sh
kubectl get node cka-sim-worker2 -o wide        # INTERNAL-IP: the address the kubelet registers (--node-ip)
ip -4 addr show eth0
ip route show default                           # default via 172.18.0.1 dev eth0
sysctl net.ipv4.ip_forward                      # = 1
```

`net.ipv4.ip_forward = 1` lets the kernel forward packets between interfaces, here between Pod veths
and the node's network. Without it, Pods can't reach anything outside their own node. kubeadm's
preflight checks fail when it's `0`, so preparing a node includes:

```sh
cat <<EOF2 > /etc/sysctl.d/k8s.conf
net.ipv4.ip_forward = 1
EOF2
sysctl --system
```

The InternalIP matters because it's what other nodes and the control plane use to reach this
node's kubelet (port 10250). On hosts with several interfaces, kubeadm's `--node-ip` decides which one.
