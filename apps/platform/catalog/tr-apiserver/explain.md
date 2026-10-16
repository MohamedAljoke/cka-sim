The API server is a **static Pod**: the kubelet on the control plane runs it straight from
`/etc/kubernetes/manifests/kube-apiserver.yaml`. When that file is wrong, `kubectl` has nothing to
talk to, so you can't use `kubectl logs` or `describe`. You debug on the node with the container
runtime instead:

```sh
crictl ps -a --name kube-apiserver                 # Exited, restarting every few seconds
crictl logs --tail 20 <container-id>               # dial tcp 127.0.0.1:2389: connection refused
ls /var/log/pods/kube-system_kube-apiserver-*/kube-apiserver/   # the same logs as files
grep etcd-servers /etc/kubernetes/manifests/kube-apiserver.yaml
grep listen-client-urls /etc/kubernetes/manifests/etcd.yaml
vim /etc/kubernetes/manifests/kube-apiserver.yaml  # --etcd-servers=https://127.0.0.1:2379
watch crictl ps --name kube-apiserver
kubectl get nodes
```

etcd serves clients on **2379** and talks to its peers on **2380**. The API server stores every
object in etcd, so it can't start without it. The kubelet notices a changed manifest within about
20 seconds, then the API server needs a little longer to become ready. Be patient instead of
editing again.

The trap is "fixing" etcd to listen on the wrong port. Make the client match the server, not the
other way round. And keep a copy before you edit (`cp kube-apiserver.yaml /root/`), never inside
`/etc/kubernetes/manifests/`, where the kubelet would run the copy too.
