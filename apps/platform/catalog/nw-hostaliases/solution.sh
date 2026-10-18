kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: legacy
  namespace: nw-hostaliases
spec:
  hostAliases:
  - ip: 10.10.0.5
    hostnames: [db.legacy.local, cache.legacy.local]
  containers:
  - name: legacy
    image: busybox:1.36
    command: [sleep, 1d]
YAML
kubectl -n nw-hostaliases wait --for=condition=Ready pod/legacy --timeout=60s
