kubectl delete priorityclass critical --ignore-not-found >/dev/null
fresh_ns wl-priority
fresh_course
alloc=$(kubectl get node cka-sim-worker2 -o jsonpath='{.status.allocatable.cpu}')
case $alloc in *m) milli=${alloc%m} ;; *) milli=$((alloc * 1000)) ;; esac
want=$((milli * 6 / 10))m
kubectl apply -f - >/dev/null <<YAML
apiVersion: apps/v1
kind: Deployment
metadata:
  name: filler
  namespace: wl-priority
spec:
  replicas: 1
  selector:
    matchLabels:
      app: filler
  template:
    metadata:
      labels:
        app: filler
    spec:
      nodeSelector:
        kubernetes.io/hostname: cka-sim-worker2
      containers:
      - name: filler
        image: busybox:1.36
        command: [sleep, 1d]
        resources:
          requests:
            cpu: $want
YAML
cat > "$COURSE/important.yaml" <<YAML
apiVersion: v1
kind: Pod
metadata:
  name: important
  namespace: wl-priority
spec:
  nodeSelector:
    kubernetes.io/hostname: cka-sim-worker2
  containers:
  - name: important
    image: busybox:1.36
    command: [sleep, 1d]
    resources:
      requests:
        cpu: $want
YAML
chmod a+rw "$COURSE/important.yaml"
kubectl -n wl-priority rollout status deployment/filler --timeout=120s >/dev/null
