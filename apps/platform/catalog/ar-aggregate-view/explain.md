The built-in ClusterRoles `view`, `edit` and `admin` are **aggregated**: their `rules` are
assembled by a controller from every ClusterRole that carries a matching label. `view` collects
everything labelled `rbac.authorization.k8s.io/aggregate-to-view: "true"`:

```sh
kubectl get clusterrole view -o jsonpath='{.aggregationRule}'
```

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: widget-viewer
  labels:
    rbac.authorization.k8s.io/aggregate-to-view: "true"
rules:
- apiGroups: [cka.example.com]
  resources: [widgets]
  verbs: [get, list, watch]
```

Within seconds, `view` contains the widget rule, and so does every binding to `view` in every
namespace. Editing `view` directly wouldn't last: the controller overwrites aggregated rules.
Operators ship labels like this with their CRDs, often `aggregate-to-edit` and `aggregate-to-admin`
roles too.

You can build your own aggregated roles with `aggregationRule.clusterRoleSelectors`. A role that
uses one has its `rules` managed for it, so leave them empty.
