NetworkPolicies follow a few rules, and every question here hangs on one of them:

1. **Isolation is per Pod and per direction.** A Pod is isolated for ingress once any policy with
   `policyTypes: [Ingress]` selects it. `default-deny` (`podSelector: {}`) selects every Pod here, so
   all ingress is denied unless a policy allows it.
2. **Policies add up.** A connection is allowed if **any** policy selecting the destination allows
   it. There is no deny rule and no order.
3. **A `podSelector` in `from` means "in the same namespace".** Pods elsewhere need a
   `namespaceSelector` (alone, or in the same `from` entry as the `podSelector` for AND).
4. **`ports` restrict the rule.** Only the listed ports are opened.

| # | Connection | Why |
|---|---|---|
| 1 | web → api:8080 | `api-from-web` allows `app=web` on 8080 |
| 2 | web → db:5432 | db only admits `app=api` |
| 3 | api → db:5432 | `db-from-api` allows it |
| 4 | api → web:80 | nothing allows ingress to `web` |
| 5 | other-ns web → api | rule 3: same-namespace `podSelector` only |
| 6 | api → db:3306 | rule 4: only 5432 is listed |

Egress isn't restricted here: no policy lists `Egress` in `policyTypes`. Remember that policies are
only as real as the network plugin. This lab enforces them with kube-router, and Calico and Cilium
do too, but a plugin that ignores them (like plain flannel) lets everything through.
