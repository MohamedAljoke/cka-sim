The kubelet doesn't call CNI plugins itself. It asks the **container runtime** to create the Pod
sandbox, and containerd's CRI plugin runs the CNI config from `conf_dir`, looking up each plugin
binary in `bin_dir`. Both live in `/etc/containerd/config.toml`:

```toml
[plugins."io.containerd.grpc.v1.cri".cni]
  bin_dir = "/opt/cni/bin"        # where ptp, host-local, portmap... are
  conf_dir = "/etc/cni/net.d"
```

```sh
kubectl -n nw-cni-bindir describe pod -l app=web | grep -A2 FailedCreatePodSandBox
#   failed to find plugin "ptp" in path [/opt/cni/plugins]
ls /opt/cni/bin/
grep -n -A3 '\.cni\]' /etc/containerd/config.toml
containerd config dump | grep -A4 '\.cni'        # the config containerd actually uses
systemctl restart containerd
```

Restarting containerd doesn't kill running containers: each one has its own shim process. That's
why the Pods that were already running kept working, since their sandboxes were set up before the
change.

The error names the path it searched, which tells you whether the config or the binaries are wrong.
Copying the binaries to the wrong path would hide the bad config until the next upgrade.
