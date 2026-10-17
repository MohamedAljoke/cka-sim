fresh_ns st-writable-layer
fresh_course
cat > "$COURSE/scratch.yaml" <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: scratch
  namespace: st-writable-layer
spec:
  containers:
  - name: scratch
    image: busybox:1.36
    command: [sh, -c, 'mkdir -p /cache; [ -f /cache/marker ] || date +%s > /cache/marker; trap "exit 0" TERM; sleep 1d & wait']
YAML
chmod a+rw "$COURSE/scratch.yaml"
kubectl apply -f "$COURSE/scratch.yaml" >/dev/null
kubectl -n st-writable-layer wait --for=condition=Ready pod/scratch --timeout=120s >/dev/null
