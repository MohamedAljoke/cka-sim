A CRD can **serve** several versions at once (clients may read and write any of them), but objects
are written to etcd in exactly one **storage** version:

```yaml
versions:
- name: v1alpha1
  served: true
  storage: false
- name: v1
  served: true
  storage: true          # exactly one version may have storage: true
```

```sh
kubectl get gadgets.v1.cka.example.com -n ar-crd-versions          # pick a version explicitly
kubectl get gadgets.v1alpha1.cka.example.com -n ar-crd-versions
kubectl get crd gadgets.cka.example.com -o jsonpath='{.status.storedVersions}'   # ["v1alpha1","v1"]
```

`status.storedVersions` lists every version that **may** still be in etcd. Old objects (like
`prototype`) stay in `v1alpha1` until something rewrites them. Before you can stop serving or
remove `v1alpha1`, migrate them (re-apply each object, or use a storage version migrator), then
remove `v1alpha1` from `storedVersions` through the status subresource.

With identical schemas, the API server converts by just changing `apiVersion` (`strategy: None`).
If the schemas differ, you need a **conversion webhook** that translates between versions.
