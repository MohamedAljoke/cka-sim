kubectl uncordon cka-sim-worker2 >/dev/null
fresh_ns tr-drain
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: node-agent
  namespace: tr-drain
spec:
  selector:
    matchLabels:
      app: node-agent
  template:
    metadata:
      labels:
        app: node-agent
    spec:
      tolerations:
      - key: dedicated
        operator: Exists
      containers:
      - name: agent
        image: busybox:1.36
        command: [sh, -c, sleep 1d]
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: tr-drain
spec:
  replicas: 2
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
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values: [cka-sim-worker2]
      tolerations:
      - key: dedicated
        operator: Exists
      containers:
      - name: api
        image: nginx:1.27-alpine
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cache
  namespace: tr-drain
spec:
  replicas: 1
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
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values: [cka-sim-worker2]
      tolerations:
      - key: dedicated
        operator: Exists
      containers:
      - name: cache
        image: busybox:1.36
        command: [sh, -c, sleep 1d]
        volumeMounts:
        - name: scratch
          mountPath: /scratch
      volumes:
      - name: scratch
        emptyDir: {}
---
apiVersion: v1
kind: Pod
metadata:
  name: debug
  namespace: tr-drain
spec:
  nodeName: cka-sim-worker2
  containers:
  - name: debug
    image: busybox:1.36
    command: [sh, -c, sleep 1d]
YAML
kubectl -n tr-drain rollout status deployment/api --timeout=120s >/dev/null
kubectl -n tr-drain rollout status deployment/cache --timeout=120s >/dev/null
kubectl -n tr-drain rollout status daemonset/node-agent --timeout=120s >/dev/null
kubectl -n tr-drain wait --for=condition=Ready pod/debug --timeout=120s >/dev/null
