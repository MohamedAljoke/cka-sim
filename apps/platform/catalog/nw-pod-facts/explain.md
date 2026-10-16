The controller-manager splits the cluster's Pod CIDR (`--cluster-cidr`, `10.244.0.0/16` here)
into one block per node and stores it in the Node's `spec.podCIDR`. The CNI's IPAM then hands out
Pod IPs from that node's block. Every Pod IP tells you which node it came from:

```sh
kubectl -n nw-pod-facts get pod tracker -o wide                     # IP and NODE columns
kubectl get node <node> -o jsonpath='{.spec.podCIDR}'
kubectl get nodes -o custom-columns=NAME:.metadata.name,CIDR:.spec.podCIDR
```

`-o wide` is the quickest view. `-o jsonpath` is what you use to write exact values to a file
without copying them by hand.
