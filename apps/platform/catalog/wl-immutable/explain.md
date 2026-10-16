`immutable: true` freezes a ConfigMap's or Secret's data: any change is rejected, and the only way
back is to delete and recreate it. In exchange the kubelet stops watching it. That saves API server
load on big clusters, and it guarantees nobody edits a config under a running app.

```sh
kubectl -n wl-immutable patch configmap app-config-v1 -p '{"immutable":true}'
kubectl -n wl-immutable edit configmap app-config-v1    # any data edit: field is immutable when `immutable` is set
```

The pattern that goes with it is **versioned names**: each change is a new ConfigMap
(`app-config-v2`), and the Deployment is pointed at it. Changing the Pod template triggers a normal
rollout, with history and `rollout undo` back to `v1`. That's also what Kustomize's hash-suffixed
generators do automatically.

Env vars are read only at container start. Even with a mutable ConfigMap, a running Pod would keep
`COLOR=green` until it was recreated.
