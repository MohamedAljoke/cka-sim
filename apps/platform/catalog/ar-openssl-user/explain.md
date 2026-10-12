The API server trusts any client certificate signed by its `--client-ca-file`. Anyone holding the
CA **key** can mint identities without asking the cluster: this is the "manual" path behind the
Certificates API.

```sh
openssl genrsa -out ops.key 2048
openssl req -new -key ops.key -subj "/CN=ops/O=ops-team" -out ops.csr      # CN = user, O = group (repeat O= for more)
openssl x509 -req -in ops.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key \
  -CAcreateserial -days 30 -out ops.crt
openssl x509 -in ops.crt -noout -subject -dates
```

Then wrap it in a kubeconfig with `kubectl config set-cluster/set-credentials/set-context
--kubeconfig <file>` and `--embed-certs`, so the file works on its own anywhere.

RBAC already grants `view` to group `ops-team`, so the new user can read as soon as the cert
carries `O=ops-team`. No binding per user is needed: groups are the way to manage many users.

Kubernetes can't revoke a client certificate: there is no CRL check. A leaked cert works until it
expires, which is why it should be short-lived and why the CA key (`ca.key`) must never leave the
control plane. Prefer the Certificates API, which needs no key access.
