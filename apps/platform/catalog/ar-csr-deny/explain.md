`kubectl get csr` shows who **sent** a request (`REQUESTOR`), not which identity it asks for. That
is inside the encoded CSR, and you must decode it before approving:

```sh
kubectl get csr
kubectl get csr ci-runner -o jsonpath='{.spec.request}' | base64 -d | openssl req -noout -subject
#   subject=CN = ci-runner, O = kubeadm:cluster-admins
kubectl get clusterrolebindings -o wide | grep kubeadm:cluster-admins    # bound to cluster-admin
kubectl certificate approve dev-anna
kubectl certificate deny ci-runner
```

The **O** fields of a client certificate become the user's groups, so a certificate is only as
harmless as the groups in it. `kubeadm:cluster-admins` is the group kubeadm's own admin kubeconfig
uses: a ClusterRoleBinding gives it `cluster-admin`. Approving `ci-runner` would hand out a
full-admin credential that Kubernetes can't revoke until it expires (there are no CRLs). The only
way to defuse it would be to change that binding for every admin.

The even stronger group `system:masters` bypasses RBAC entirely. The API server now refuses to
accept `kube-apiserver-client` CSRs asking for it at all, which is why attackers aim for groups like
this one instead. Treat approval rights on CSRs (the `certificatesigningrequests/approval`
subresource) with the same care as cluster-admin itself.
