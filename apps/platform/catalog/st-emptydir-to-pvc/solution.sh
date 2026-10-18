kubectl apply -f - <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: notes-data
  namespace: st-emptydir-to-pvc
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: standard
  resources:
    requests:
      storage: 100Mi
YAML
kubectl -n st-emptydir-to-pvc patch deployment notes --type=json -p '[{"op":"replace","path":"/spec/template/spec/volumes/0","value":{"name":"data","persistentVolumeClaim":{"claimName":"notes-data"}}}]'
kubectl -n st-emptydir-to-pvc rollout status deployment notes --timeout=120s
kubectl -n st-emptydir-to-pvc delete pod -l app=notes
kubectl -n st-emptydir-to-pvc rollout status deployment notes --timeout=120s
until [ "$(kubectl -n st-emptydir-to-pvc get pods -l app=notes --field-selector status.phase=Running -o name | wc -l)" = 1 ]; do sleep 2; done
kubectl -n st-emptydir-to-pvc exec deploy/notes -- cat /data/notes.txt
