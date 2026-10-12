Two commands answer "who" and "what":

```sh
kubectl auth whoami                                   # the identity the API server sees for your credentials
#   Username  kubernetes-admin
#   Groups    [kubeadm:cluster-admins system:authenticated]
kubectl auth can-i --list -n ar-whoami --as=system:serviceaccount:ar-whoami:viewer
#   Resources    Non-Resource URLs  Resource Names  Verbs
#   configmaps   []                 []              [get list]
#   selfsubjectreviews.authentication.k8s.io  []  []  [create]   <- granted to everyone
```

`whoami` (a SelfSubjectReview) is the reliable way to learn your identity: a kubeconfig's user
*entry name* is only a label. kubeadm's admin user is in group `kubeadm:cluster-admins`, which a
ClusterRoleBinding ties to `cluster-admin`. That binding is what makes it all-powerful.

`--as` (and `--as-group`) **impersonate** another identity. It needs the `impersonate` verb, which
admins have. Combined with `can-i`, it's the standard way to test RBAC changes for a user or
ServiceAccount without its credentials. ServiceAccounts are users named
`system:serviceaccount:<namespace>:<name>`, in groups `system:serviceaccounts` and
`system:serviceaccounts:<namespace>`.
