**Pod Security Admission** is built into the API server and replaced PodSecurityPolicy. You turn it
on per namespace with labels, choosing one of three standards (`privileged`, `baseline`,
`restricted`) for each of three modes:

| Label | Effect |
|---|---|
| `pod-security.kubernetes.io/enforce=<level>` | reject Pods that break the standard |
| `pod-security.kubernetes.io/warn=<level>` | allow, but warn the client |
| `pod-security.kubernetes.io/audit=<level>` | allow, but note it in the audit log |

`restricted` requires, on every container: `runAsNonRoot: true` (with a non-zero user if the image
runs as root), `allowPrivilegeEscalation: false`, `capabilities.drop: [ALL]`, and
`seccompProfile.type: RuntimeDefault` (or `Localhost`).

```sh
kubectl label ns wl-psa pod-security.kubernetes.io/enforce=restricted pod-security.kubernetes.io/enforce-version=latest
#   Warning: existing pods in namespace "wl-psa" violate the new PodSecurity enforce level
kubectl -n wl-psa rollout restart deploy web       # the new Pods are rejected
kubectl -n wl-psa describe rs <new-rs>             # FailedCreate: violates PodSecurity "restricted:latest": ...
```

Enforcement only applies when Pods are **created**. Labelling a namespace never touches running
Pods; it only warns about them. A Deployment's rejected Pods show up as `FailedCreate` events on its
ReplicaSet, not on the Deployment. The error lists exactly which fields to fix.
