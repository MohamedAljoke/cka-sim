A kubeconfig has three lists: **clusters** (server and CA), **users** (credentials) and
**contexts** (a cluster plus a user plus a default namespace), and one `current-context`. Every
`kubectl config set-*` command takes `--kubeconfig <file>` to build a new file instead of editing
`~/.kube/config`.

```sh
KC=/opt/course/ar-sa-kubeconfig/ci-bot.kubeconfig
kubectl config set-cluster cka-sim --server=https://cka-sim-control-plane:6443 \
  --certificate-authority=/etc/kubernetes/pki/ca.crt --embed-certs --kubeconfig $KC
kubectl config set-credentials ci-bot --token="$(kubectl -n ar-sa-kubeconfig create token ci-bot --duration=1h)" --kubeconfig $KC
kubectl config set-context ci-bot --cluster=cka-sim --user=ci-bot --namespace=ar-sa-kubeconfig --kubeconfig $KC
kubectl config use-context ci-bot --kubeconfig $KC
kubectl --kubeconfig $KC auth whoami
```

`kubectl create token` asks the TokenRequest API for a **short-lived, bound** token. It isn't
stored anywhere in the cluster and expires on its own, which is what you want to hand to CI.
`--embed-certs` copies the CA into the file, so it works on another machine; without it, the file
only holds a path.

The ServiceAccount's RBAC still applies, so the token can do exactly what `ci-bot` may do.
