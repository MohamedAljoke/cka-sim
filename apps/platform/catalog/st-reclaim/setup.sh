fresh_ns st-reclaim
for pv in $(kubectl get pv -o jsonpath='{range .items[?(@.spec.claimRef.namespace=="st-reclaim")]}{.metadata.name}{" "}{end}'); do
  kubectl delete pv "$pv" --wait=false >/dev/null
done
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: orders
  namespace: st-reclaim
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: standard
  resources:
    requests:
      storage: 100Mi
---
apiVersion: v1
kind: Pod
metadata:
  name: orders-db
  namespace: st-reclaim
spec:
  containers:
  - name: db
    image: busybox:1.36
    command: [sh, -c, 'echo order-1 >> /data/orders; sleep 1d']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: orders
YAML
kubectl -n st-reclaim wait --for=condition=Ready pod/orders-db --timeout=120s >/dev/null
