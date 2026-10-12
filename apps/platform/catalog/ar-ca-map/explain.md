kubeadm creates **three** CAs, each trusted for one job:

| CA | Files | Signs | Trusted by |
|---|---|---|---|
| cluster CA | `ca.crt` / `ca.key` | API server serving cert, API server→kubelet client cert, admin/controller-manager/scheduler kubeconfigs, kubelet client certs | API server (`--client-ca-file`), kubelets, everyone's kubeconfig |
| etcd CA | `etcd/ca.crt` / `etcd/ca.key` | etcd server and peer certs, `apiserver-etcd-client.crt`, `etcd/healthcheck-client.crt` | etcd (`--trusted-ca-file`) and the API server (`--etcd-cafile`) |
| front-proxy CA | `front-proxy-ca.crt` / `.key` | `front-proxy-client.crt`, which the API server uses toward aggregated APIs such as metrics-server | the aggregated API servers (`--requestheader-client-ca-file`) |

```sh
openssl x509 -in /etc/kubernetes/pki/apiserver-etcd-client.crt -noout -issuer      # CN = etcd-ca
openssl verify -CAfile /etc/kubernetes/pki/etcd/ca.crt /etc/kubernetes/pki/apiserver-etcd-client.crt
kubeadm certs check-expiration                   # the CERTIFICATE AUTHORITY column says it too
```

Separate CAs limit the damage of a leak: a cert from the cluster CA can't talk to etcd directly,
and etcd only trusts its own CA. `etcdctl` needs etcd-CA files, never the cluster CA's. That's the
usual mistake in backup tasks.
