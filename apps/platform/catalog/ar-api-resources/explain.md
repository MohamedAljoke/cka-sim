Every kind lives in an **API group** at a **version**. The core group has no name, so its
apiVersion is just `v1` (Pods, Services, ConfigMaps). Everything else is `<group>/<version>`:
`apps/v1`, `coordination.k8s.io/v1`, `scheduling.k8s.io/v1`, `storage.k8s.io/v1`. Resources are
either **namespaced** or **cluster-scoped**.

```sh
kubectl api-resources                          # NAME SHORTNAMES APIVERSION NAMESPACED KIND
kubectl api-resources --namespaced=false       # only cluster-scoped kinds
kubectl api-resources --api-group=storage.k8s.io
kubectl api-versions                           # every group/version the server serves
kubectl explain priorityclass | head -3        # GROUP and VERSION at the top
```

Two traps: `events` exists twice, as core `v1` events and as `events.k8s.io/v1`, and
`kubectl api-resources` lists the core one first. Also, some names don't say their scope:
StorageClasses, PriorityClasses and CSINodes are cluster-scoped, while Leases are namespaced.

The group matters in RBAC (`apiGroups: ["apps"]` for Deployments, `[""]` for core kinds) and in
every manifest's `apiVersion`.
