Some actions are **subresources** of an object, with their own RBAC entries written
`<resource>/<subresource>`:

| Action | Resource | Verb |
|---|---|---|
| `kubectl logs` | `pods/log` | `get` |
| `kubectl exec`, `kubectl cp` | `pods/exec` | `create` |
| `kubectl port-forward` | `pods/portforward` | `create` |
| `kubectl scale` | `deployments/scale` (group `apps`) | `get`, `update` / `patch` |
| `kubectl attach` | `pods/attach` | `create` |

```sh
kubectl -n ar-subresources create role support --verb=get,list --resource=pods,pods/log   # then add the rest in YAML
kubectl auth can-i create pods --subresource=exec -n ar-subresources --as=x --as-group=support
kubectl auth can-i update deployments --subresource=scale -n ar-subresources --as=x --as-group=support
```

Rights on the parent don't cover its subresources, and the reverse is true too. That's what makes
"may scale but not edit" possible: `deployments/scale` only touches `replicas`, while `update` on
`deployments` could change the image. Be careful with `pods/exec`, though: a shell in a Pod carries
everything that Pod can reach, including its ServiceAccount token.
