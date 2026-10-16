`policy-c.yaml` is the answer: `backend` Pods in the same Namespace, on 5432 only.

| Policy | Rule | Why it's wrong |
|---|---|---|
| a | `podSelector: app=backend` | no `ports`, so every port is open, `9000` too |
| b | `namespaceSelector: {}` **and** `podSelector: app=backend` in one entry | `{}` matches every Namespace, so `backend` Pods anywhere get in |
| c | `podSelector: app=backend`, port 5432 | meets the requirement and nothing more |
| d | `podSelector: {}`, port 5432 | every Pod in the Namespace, `frontend` too |

Applying more than one doesn't help. Policies add up, so the most permissive one wins. To compare
candidates quickly, read each `from` entry as one sentence, with "and" inside an entry and "or"
between entries:

```bash
kubectl apply -f /opt/course/nw-netpol-pick/policy-c.yaml
kubectl -n nw-netpol-pick describe networkpolicy db-c
```

`kubectl describe networkpolicy` prints the rules in that same plain-language form.
