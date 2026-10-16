A CRD's name is always `<plural>.<group>`, so a group's CRDs are easy to filter. Custom resources
then behave like any other kind, including `-A` and output formatting:

```sh
kubectl get crd | grep monitoring.cka.io
kubectl api-resources --api-group=monitoring.cka.io       # kinds, short names, namespaced or not
kubectl get dashboards.monitoring.cka.io -A
kubectl get dashboards -A -o custom-columns=NS:.metadata.namespace,NAME:.metadata.name --no-headers
```

Use the full `<plural>.<group>` name when two groups could share a plural. Here
`panels.grafana.cka.io` exists too, and on real clusters names like `certificates` exist in
several groups. `grep dashboard` alone would also catch the Panel named `dashboard-cpu`. Filter
on the group, not on a word.
