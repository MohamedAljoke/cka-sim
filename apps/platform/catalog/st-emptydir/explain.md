An `emptyDir` is created empty when the Pod is scheduled to a node, and deleted when the Pod is
removed. Every container in the Pod that mounts it sees the same files, which makes it the usual
way for containers in one Pod to share data. It survives container restarts, not Pod deletion.

```yaml
volumes:
- name: data
  emptyDir: {}                                  # on the node's disk (under /var/lib/kubelet/pods/)
- name: fast
  emptyDir: {medium: Memory, sizeLimit: 64Mi}   # a tmpfs in RAM
```

```sh
kubectl -n st-emptydir exec share -c reader -- tail -3 /data/out.txt
kubectl -n st-emptydir exec share-mem -c reader -- df -h /data      # tmpfs
```

A memory-backed `emptyDir` is fast and never touches disk, which suits scratch space and secrets.
But what you write there **counts against the container's memory limit**. Without a
`sizeLimit`, a runaway writer can get the Pod OOM-killed or evicted. A disk-backed `emptyDir`
over its `sizeLimit` gets the Pod evicted too.
