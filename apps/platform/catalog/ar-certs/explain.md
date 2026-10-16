kubeadm creates the cluster's certificates under `/etc/kubernetes/pki/`. Leaf certs (API server,
kubelet client, etcd clients, the `*.conf` kubeconfigs) last **1 year**, and the CAs last 10 years.
`kubeadm upgrade` renews leaf certs automatically. A cluster that is never upgraded stops working
after a year.

```sh
kubeadm certs check-expiration                       # EXPIRES and RESIDUAL TIME per cert, CAs at the bottom
openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -enddate
kubeadm certs renew apiserver                        # or: kubeadm certs renew all
openssl x509 -in /etc/kubernetes/pki/apiserver.crt -noout -dates
```

Renewing only rewrites the files. The running API server loaded its certificate at startup, so it
keeps serving the old one until it restarts. Restart the static Pod:

```sh
crictl stop $(crictl ps -q --name '^kube-apiserver$')          # the kubelet starts it again
# or: mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp/; sleep 20; mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests/
openssl s_client -connect 127.0.0.1:6443 </dev/null 2>/dev/null | openssl x509 -noout -dates
```

The trap is stopping after `renew` and believing it's done: the restart is what makes it take
effect. `kubectl delete pod kube-apiserver-...` doesn't restart a static Pod. It only deletes the
mirror object.
