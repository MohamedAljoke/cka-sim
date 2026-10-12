`hostPort` is Kubernetes' version of `docker run -p 8081:80`: the CNI's **portmap** plugin adds a
DNAT rule on that one node, sending `<node-ip>:8081` to the Pod's port 80. The Pod keeps its own
network namespace and IP.

```yaml
ports:
- containerPort: 80
  hostPort: 8081
```

```sh
curl http://<worker2-ip>:8081
iptables -t nat -S | grep 8081          # on the node: CNI-HOSTPORT-DNAT rules
```

Compare the three ways to reach a Pod from outside the cluster network:

| | Opens the port on | Survives the Pod moving | Port range |
|---|---|---|---|
| `hostPort` | only the node running the Pod | no: the address changes with the node | any |
| `hostNetwork` | that node, sharing its whole network | no | any |
| NodePort Service | every node | yes | 30000–32767 |

Every Pod with the same `hostPort` needs a different node, and the scheduler enforces that. It's
used by ingress controllers deployed as DaemonSets. Otherwise prefer a Service. `hostPort` needs a
CNI that supports portmap; kindnet chains it in its config.
