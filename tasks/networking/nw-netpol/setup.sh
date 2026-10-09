fresh_ns payments
fresh_ns monitoring
k apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata: {name: api, namespace: payments}
spec:
  replicas: 1
  selector: {matchLabels: {app: api}}
  template:
    metadata: {labels: {app: api}}
    spec:
      containers: [{name: nginx, image: "nginx:1.27-alpine", ports: [{containerPort: 80}]}]
---
apiVersion: v1
kind: Service
metadata: {name: api, namespace: payments}
spec:
  selector: {app: api}
  ports: [{port: 80}]
---
apiVersion: v1
kind: Pod
metadata: {name: frontend, namespace: payments, labels: {app: frontend}}
spec:
  containers: [{name: shell, image: "busybox:1.36", command: [sleep, "86400"]}]
---
apiVersion: v1
kind: Pod
metadata: {name: batch, namespace: payments, labels: {app: batch}}
spec:
  containers: [{name: shell, image: "busybox:1.36", command: [sleep, "86400"]}]
---
apiVersion: v1
kind: Pod
metadata: {name: prober, namespace: monitoring, labels: {app: prober}}
spec:
  containers: [{name: shell, image: "busybox:1.36", command: [sleep, "86400"]}]
YAML
k -n payments rollout status deploy/api --timeout=180s >/dev/null
k -n payments wait pod/frontend pod/batch --for=condition=Ready --timeout=180s >/dev/null
k -n monitoring wait pod/prober --for=condition=Ready --timeout=180s >/dev/null
