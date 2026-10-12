kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: node-agent
  namespace: nw-hostnetwork
spec:
  hostNetwork: true
  nodeSelector:
    kubernetes.io/hostname: cka-sim-worker2
  containers:
  - name: agent
    image: busybox:1.36
    command: [sh, -c, 'mkdir -p /www && echo agent > /www/index.html && httpd -f -p 18080 -h /www']
YAML
kubectl -n nw-hostnetwork wait --for=condition=Ready pod/node-agent --timeout=60s
