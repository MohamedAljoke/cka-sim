A Namespace is a name scope: two Pods can both be called `web` if they live in different Namespaces,
and RBAC, ResourceQuotas and most NetworkPolicies apply per Namespace. Every namespaced command works in
your current Namespace (`default`) unless you pass `-n`, which is the usual exam slip: the object gets
created, just in the wrong place.

```sh
kubectl create namespace galaxy
kubectl label namespace galaxy env=training
kubectl -n galaxy run probe --image=nginx:1.27
kubectl -n galaxy create deployment web --image=nginx:1.27 --replicas=2
kubectl get pods -A --field-selector metadata.name=hermes
```

`-A` (`--all-namespaces`) searches every Namespace; its first column is the Namespace.
To stop typing `-n`, switch the context's default: `kubectl config set-context --current --namespace=galaxy`,
and switch back afterwards. Nodes, PersistentVolumes, ClusterRoles and Namespaces themselves are not
namespaced: `kubectl api-resources --namespaced=false` lists them.
