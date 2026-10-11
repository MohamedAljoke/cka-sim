**Concept.** RBAC is three objects: **who** (a ServiceAccount, User or Group), **what**
(a Role: verbs on resources, scoped to one namespace — or a ClusterRole for the whole cluster)
and the **link** (a RoleBinding). Nothing is allowed unless a binding grants it; there are no
deny rules, so "least privilege" means listing only the verbs you need.

**Imperative is faster on the exam:**
`k create role deployer --verb=get,list,create,update --resource=deployments -n ci` — kubectl
resolves `deployments` to the `apps` API group for you.

**Verify, don't assume:** `k auth can-i <verb> <resource> -n <ns> --as=system:serviceaccount:<ns>:<name>`.
The `--as` username format for ServiceAccounts is worth memorising. `can-i` exits non-zero on
`no`, which is why the solution ends with `|| true`.
