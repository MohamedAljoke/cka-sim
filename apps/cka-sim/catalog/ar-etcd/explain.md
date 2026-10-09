**Concept.** etcd *is* the cluster state: every object you `kubectl apply` ends up as a key in
etcd. A snapshot is a point-in-time copy of that keyspace; restoring it rewinds the whole cluster.

**Backup** needs etcd's client TLS material — read it from the static Pod, don't memorise it:
`grep -E 'cert|key|trusted-ca|listen-client' /etc/kubernetes/manifests/etcd.yaml`, then
`etcdctl --endpoints=https://127.0.0.1:2379 --cacert=… --cert=… --key=… snapshot save <file>`.
(etcdctl 3.6 speaks v3 only — the old `ETCDCTL_API=3` prefix is no longer needed.)

**Restore** is offline and local — no endpoints, no certs:
`etcdutl snapshot restore <file> --data-dir /var/lib/etcd-restore` creates a fresh data dir.
Then edit `/etc/kubernetes/manifests/etcd.yaml` so the `etcd-data` **hostPath** points at it
(`--data-dir` and `mountPath` stay `/var/lib/etcd` — that is the path *inside* the container).
The kubelet restarts etcd from the new data. Never restore into the directory etcd is using.

**The step most guides skip — and the one this task grades.** The API server keeps a *watch
cache* of etcd. Swap etcd underneath a running API server and it keeps serving the old,
pre-restore objects even though etcd holds the restored ones (`etcdctl get /registry/… --prefix
--keys-only` shows the truth). The Kubernetes docs' procedure: stop the API server (move
`kube-apiserver.yaml` out of `manifests/`), restore, then bring it back. Restoring with
`--bump-revision 1000000000 --mark-compacted` additionally forces every watcher to re-list
instead of trusting what it cached.
