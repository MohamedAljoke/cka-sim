# kubectl -n st-pending-pvc describe pvc reports -> no persistent volumes available for this claim
# kubectl get pv pv-reports                      -> 1Gi RWO manual; the claim asks for 2Gi RWX
kubectl -n st-pending-pvc get pod reporter -o yaml > /tmp/reporter.yaml
kubectl -n st-pending-pvc delete pod reporter --grace-period=1
kubectl -n st-pending-pvc delete pvc reports
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: reports
  namespace: st-pending-pvc
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: manual
  resources:
    requests:
      storage: 1Gi
YAML
kubectl apply -f /tmp/reporter.yaml
kubectl -n st-pending-pvc wait --for=condition=Ready pod/reporter --timeout=90s
