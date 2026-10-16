Nodes and PersistentVolumes are **cluster-scoped**: they belong to no namespace, so a Role (which
lives in a namespace) can't grant access to them. You need a **ClusterRole** for the rules and a
**ClusterRoleBinding** to grant them everywhere.

```sh
kubectl create clusterrole node-viewer --verb=get,list --resource=nodes,persistentvolumes
kubectl create clusterrolebinding node-viewer --clusterrole=node-viewer --group=auditors
kubectl auth can-i list persistentvolumes --as=jane --as-group=auditors   # yes
kubectl auth can-i delete nodes --as=jane --as-group=auditors             # no
```

Kubernetes has no group objects. A group is just a name the authenticator hands over, for example
the `O=` field of a client certificate. To test a group with `--as-group`, you also need `--as`
with any user name.

The trap is `--verb='*'` or adding `watch` "to be safe": least privilege means exactly the verbs
asked for. `kubectl get pv -w` needs `watch`, but plain `get` and `list` don't.
