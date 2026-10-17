fresh_ns wl-pending
kubectl label nodes --all zone- >/dev/null 2>&1 || true
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: wl-pending
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: zone
                operator: In
                values: [moon]
      containers:
      - name: nginx
        image: nginx:1.27
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: batch
  namespace: wl-pending
spec:
  replicas: 1
  selector:
    matchLabels:
      app: batch
  template:
    metadata:
      labels:
        app: batch
    spec:
      containers:
      - name: nginx
        image: nginx:1.27
        resources:
          requests:
            memory: 1Ti
YAML
unscheduled() { eq False "$(kubectl -n wl-pending get pods -l app="$1" -o jsonpath='{.items[0].status.conditions[?(@.type=="PodScheduled")].status}')"; }
wait_for 60 unscheduled api
wait_for 60 unscheduled batch
