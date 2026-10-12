When the API server is down, `kubectl logs` can't help: it asks the API server for the logs. The
kubelet and the container runtime still work, so you read logs on the node:

- `crictl ps -a` lists containers, including exited ones. `crictl logs <id>` prints a
  container's output.
- The same output is kept as files under `/var/log/pods/<namespace>_<pod>_<uid>/<container>/`.
- `journalctl -u kubelet` tells you what the kubelet thinks, but not why the container died.

```sh
crictl ps -a | grep -E 'etcd|kube-apiserver'
crictl logs $(crictl ps -a -q --name '^etcd$' | head -1) 2>&1 | tail -5
ls /var/log/pods/kube-system_etcd-*/etcd/
ls /etc/kubernetes/pki/etcd/                   # server.crt, not server.cert
vim /etc/kubernetes/manifests/etcd.yaml
```

Here the API server crashes too, but only because it can't reach etcd. Its log says
`connection refused` to 127.0.0.1:2379. The root cause is the component it depends on. Follow
the dependency: kubectl → API server → etcd. etcd writes its fatal error at the end of its log,
just before it exits.

Static Pods restart with back-off. After fixing the manifest, give the kubelet up to a minute to
start etcd, and then the API server reconnects on its own.
