kubectl taint node cka-sim-worker dedicated=gpu:NoSchedule
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: gpu-job
  namespace: wl-taints
spec:
  nodeSelector:
    kubernetes.io/hostname: cka-sim-worker
  tolerations:
  - key: dedicated
    operator: Equal
    value: gpu
    effect: NoSchedule
  containers:
  - name: nginx
    image: nginx:1.27
YAML
kubectl -n wl-taints run regular --image=nginx:1.27
kubectl -n wl-taints wait pod gpu-job regular --for=condition=Ready --timeout=120s
