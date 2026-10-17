fresh_ns wl-host-process
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: proc
  namespace: wl-host-process
spec:
  nodeName: cka-sim-control-plane
  securityContext:
    runAsUser: 3000
  containers:
  - name: proc
    image: busybox:1.36
    command: [sleep, "86401"]
YAML
kubectl -n wl-host-process wait --for=condition=Ready pod/proc --timeout=120s >/dev/null
