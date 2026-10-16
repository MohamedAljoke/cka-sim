A kubeconfig's `client-certificate-data` is the PEM certificate, base64-encoded once more. Decode it
and read it with openssl:

```sh
kubectl --kubeconfig auditor.kubeconfig config view --raw -o jsonpath='{.users[0].user.client-certificate-data}' | base64 -d > auditor.crt
openssl x509 -in auditor.crt -noout -subject -issuer -dates
#   subject=O = auditors, O = readonly, CN = ext-auditor
#   notAfter=Dec 26 10:03:11 2026 GMT
```

For the API server, CN is the **user** and every O is a **group**. RBAC bindings to either apply to
whoever holds this file. `--raw` is needed because `config view` otherwise prints `DATA+OMITTED`.
`-nameopt RFC2253` prints the subject as `CN=...,O=...`, which is easy to split.

Since a client certificate can't be revoked, the way to cut off a leaked one is to remove the
RBAC bindings for its user and groups, and wait for the expiry date you just found.
