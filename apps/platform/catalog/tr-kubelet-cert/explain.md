The kubelet logs in to the API server with a client certificate. Its kubeconfig
`/etc/kubernetes/kubelet.conf` points at `/var/lib/kubelet/pki/kubelet-client-current.pem`, a
symlink to the dated cert and key that the kubelet rotates by itself. The API server only trusts
client certs signed by its `--client-ca-file` (the cluster CA). A cert with the right name,
`CN=system:node:cka-sim-worker, O=system:nodes`, but the wrong signer is rejected as
`Unauthorized`, so the node stops posting status and goes `NotReady`.

```sh
journalctl -u kubelet --no-pager | grep -i -m3 unauthorized
grep client- /etc/kubernetes/kubelet.conf
ls -l /var/lib/kubelet/pki/
openssl x509 -in /var/lib/kubelet/pki/kubelet-client-current.pem -noout -subject -issuer -dates
openssl verify -CAfile /etc/kubernetes/pki/ca.crt /var/lib/kubelet/pki/kubelet-client-2*.pem
ln -sfn /var/lib/kubelet/pki/<the file that verifies> /var/lib/kubelet/pki/kubelet-client-current.pem
systemctl restart kubelet
```

The trap is trusting the subject: the forged cert has exactly the right CN and O. The **issuer**
is what matters. Compare it with `openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -subject`.

If no valid cert were left, you would issue a new one through the Certificates API (a CSR with
signer `kubernetes.io/kube-apiserver-client-kubelet`, approved on the control plane) or rejoin
the node with a bootstrap token.
