A **hostPath** volume mounts a path from the node into the Pod. Its optional `type` is a
precondition the kubelet checks before mounting:

| type | Meaning |
|---|---|
| `""` (unset) | no check: mount whatever is there, creating nothing |
| `Directory` / `File` / `Socket` / `CharDevice` / `BlockDevice` | it must already exist as that kind |
| `DirectoryOrCreate` | create an empty directory (0755, owned by the kubelet's user) if missing |
| `FileOrCreate` | create an empty file if missing (but not its parent directory) |

When the check fails, the Pod stays `ContainerCreating` and the reason is only in the events:

```sh
kubectl -n st-hostpath-type describe pod agent | tail -3
#   MountVolume.SetUp failed for volume "state" : hostPath type check failed: /var/lib/agent-data is not a directory
```

hostPath ties a Pod to one node's disk: a Pod moved to another node sees a different directory.
It's meant for node agents (log shippers, CNI, CSI drivers) that really need host files, not for
app data. That's also why `restricted` Pod Security forbids it.
