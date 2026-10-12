fresh_ns st-wffc
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cache
  namespace: st-wffc
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: standard
  resources:
    requests:
      storage: 50Mi
YAML
waiting() { kubectl -n st-wffc get events --field-selector involvedObject.name=cache,reason=WaitForFirstConsumer -o name | grep -q .; }
wait_for 60 waiting
