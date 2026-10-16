etcd runs as a static Pod, so its settings are flags in `/etc/kubernetes/manifests/etcd.yaml`.
Talking to it with `etcdctl` needs the endpoint and three TLS files, because etcd only accepts
clients with a certificate signed by the etcd CA:

```sh
grep -E 'data-dir|listen-client|cert-file|key-file|trusted-ca' /etc/kubernetes/manifests/etcd.yaml
export ETCDCTL_CACERT=/etc/kubernetes/pki/etcd/ca.crt \
       ETCDCTL_CERT=/etc/kubernetes/pki/etcd/server.crt \
       ETCDCTL_KEY=/etc/kubernetes/pki/etcd/server.key
etcdctl --endpoints=https://127.0.0.1:2379 endpoint status -w table   # VERSION, DB SIZE, IS LEADER
etcdctl --endpoints=https://127.0.0.1:2379 member list -w table       # NAME, PEER and CLIENT ADDRS
```

`etcdctl version` prints the **client's** version, which can differ from the server's. The
`endpoint status` VERSION column is the server. The data dir (`/var/lib/etcd`) is mounted into the
Pod with a hostPath volume, and it's what you point `etcdutl snapshot restore --data-dir` at
during a restore.

`-w table`, `-w json` and `-w simple` change the output format. `json` plus `jq` is the most
reliable for writing answer files.
