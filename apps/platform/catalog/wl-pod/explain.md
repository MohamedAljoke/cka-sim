A Pod is the smallest thing Kubernetes runs: one or more containers that share a network and storage.
`kubectl run` creates a single Pod with no controller behind it, so nothing recreates it if it is deleted.

```sh
kubectl -n wl-pod run web --image=nginx:1.27
kubectl -n wl-pod get pod web
```

Add `--dry-run=client -o yaml > pod.yaml` to get a manifest to edit instead of creating the Pod at once.
