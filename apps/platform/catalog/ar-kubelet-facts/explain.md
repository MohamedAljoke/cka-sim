The kubelet is a systemd service, so its startup comes from the unit file and its drop-ins.
`systemctl cat kubelet` shows them all at once. On kubeadm nodes the drop-in
`10-kubeadm.conf` passes `--config=/var/lib/kubelet/config.yaml`, a `KubeletConfiguration` that
holds almost every setting, and `--kubeconfig=/etc/kubernetes/kubelet.conf` for talking to the
API server.

```sh
systemctl cat kubelet | grep -E -- '--config|--kubeconfig'
grep -E 'staticPodPath|cgroupDriver|containerRuntimeEndpoint|clusterDNS|clusterDomain' /var/lib/kubelet/config.yaml
ps -o args= -C kubelet | tr ' ' '\n' | grep -- --        # the flags the running process got
```

The **cgroup driver** must match the container runtime's (`SystemdCgroup = true` in
`/etc/containerd/config.toml` means `systemd`). A mismatch makes Pods fail to start on that node.
**staticPodPath** is the folder the kubelet watches for static Pod manifests. On workers it's
usually empty, but it exists.

Change a setting by editing the config file, then `systemctl restart kubelet`. Flags in
`/var/lib/kubelet/kubeadm-flags.env` win over the file when both are set.
