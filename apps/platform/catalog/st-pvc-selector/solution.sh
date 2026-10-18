kubectl apply -f - <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: fast
  namespace: st-pvc-selector
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: tiered
  selector:
    matchLabels:
      disk: ssd
  resources:
    requests:
      storage: 1Gi
YAML
kubectl -n st-pvc-selector wait --for=jsonpath='{.status.phase}'=Bound pvc/fast --timeout=30s
kubectl -n st-pvc-selector get pvc fast -o jsonpath='{.status.capacity.storage}' > /opt/course/st-pvc-selector/capacity.txt
