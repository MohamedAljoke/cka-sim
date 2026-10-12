The kubelet doesn't run containers itself. It talks to a container runtime over the **CRI**, a
gRPC API on a Unix socket. Which socket is set by `containerRuntimeEndpoint` in the kubelet config
file (`/var/lib/kubelet/config.yaml` on kubeadm nodes; older clusters used the
`--container-runtime-endpoint` flag). If the socket doesn't exist, the kubelet exits at startup,
systemd restarts it in a loop, and the node goes `NotReady`.

```sh
systemctl status kubelet
journalctl -u kubelet --no-pager | tail -20      # ... dial unix /var/run/cri-dockerd.sock: no such file
ls -l /run/containerd/containerd.sock
crictl info >/dev/null && echo runtime ok         # crictl reads /etc/crictl.yaml, not the kubelet config
vim /var/lib/kubelet/config.yaml                  # containerRuntimeEndpoint: unix:///run/containerd/containerd.sock
systemctl restart kubelet
kubectl get node cka-sim-worker2
```

The trap is checking containerd first: `systemctl status containerd` is fine and `crictl ps` works,
because the runtime is healthy. Only the kubelet is pointed at the wrong place. Editing the file
is what survives restarts; a flag passed by hand would not.
