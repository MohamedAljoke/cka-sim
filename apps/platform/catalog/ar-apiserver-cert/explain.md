A TLS client accepts a server only if the name it connected to is in the certificate's **Subject
Alternative Names** (the CN no longer counts). The API server's cert lists every name and IP
clients may use: the node's names, `kubernetes`, `kubernetes.default.svc.cluster.local`, the
Service IP `10.96.0.1`, and the node IP.

```sh
openssl x509 -in /etc/kubernetes/pki/apiserver.crt -noout -text | grep -A1 'Subject Alternative Name'
openssl x509 -in /etc/kubernetes/pki/apiserver.crt -noout -issuer -subject -dates
openssl s_client -connect 127.0.0.1:6443 </dev/null 2>/dev/null | openssl x509 -noout -ext subjectAltName   # what's really served
kubeadm certs check-expiration
```

A client using a missing name fails with `x509: certificate is valid for ..., not k8s.cka.internal`.
The fix is to add the name to kubeadm's `apiServer.certSANs`, delete `apiserver.crt` and `.key`,
run `kubeadm init phase certs apiserver --config <file>`, and restart the API server.

`-noout -text` prints everything. `-issuer`, `-subject`, `-dates` and `-ext subjectAltName` print
just one part, which is handier for answer files.
