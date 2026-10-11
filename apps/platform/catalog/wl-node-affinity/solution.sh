kubectl label node cka-sim-worker2 disktype=ssd
kubectl -n wl-node-affinity run fast --image=nginx:1.27 --overrides='{"spec":{"nodeSelector":{"disktype":"ssd"}}}'
# kubectl -n wl-node-affinity create deployment cache --image=nginx:1.27 --replicas=2 --dry-run=client -o yaml > cache.yaml
kubectl apply -f - <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cache
  namespace: wl-node-affinity
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cache
  template:
    metadata:
      labels:
        app: cache
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: disktype
                operator: In
                values: [ssd, nvme]
      containers:
      - name: nginx
        image: nginx:1.27
YAML
kubectl -n wl-node-affinity wait pod fast --for=condition=Ready --timeout=120s
kubectl -n wl-node-affinity rollout status deployment/cache --timeout=120s
