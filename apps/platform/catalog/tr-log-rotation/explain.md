Container stdout and stderr end up as files under `/var/log/pods/<ns>_<pod>_<uid>/<container>/`,
written by the container runtime. The **kubelet** rotates them, using two settings in its config
file:

```yaml
containerLogMaxSize: 5Mi      # rotate when the current file reaches this size (default 10Mi)
containerLogMaxFiles: 3       # keep at most this many files per container (default 5)
```

```sh
vim /var/lib/kubelet/config.yaml
systemctl restart kubelet
kubectl get --raw /api/v1/nodes/cka-sim-worker2/proxy/configz | jq '.kubeletconfig | {containerLogMaxSize, containerLogMaxFiles}'
ls -l /var/log/pods/*/*/                                       # 0.log, 0.log.<timestamp>...
```

Rotated logs are what `kubectl logs` can no longer show: it reads only the current file. For
history beyond that, clusters ship logs off the node with an agent (a DaemonSet reading
`/var/log/pods`).

Settings in the config file only take effect after `systemctl restart kubelet`. The `configz`
endpoint shows what the running kubelet actually loaded.
