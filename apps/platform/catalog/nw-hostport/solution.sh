kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: web
  namespace: nw-hostport
spec:
  nodeSelector:
    kubernetes.io/hostname: cka-sim-worker2
  containers:
  - name: web
    image: nginx:1.27-alpine
    ports:
    - containerPort: 80
      hostPort: 8081
YAML
kubectl -n nw-hostport wait --for=condition=Ready pod/web --timeout=60s
