etcd_tools
fresh_course
fresh_ns etcd-check
kubectl -n etcd-check create configmap marker --from-literal=state=from-backup >/dev/null
etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key \
  snapshot save "$COURSE/etcd-backup-old.db" >/dev/null
kubectl -n etcd-check delete configmap marker >/dev/null
kubectl -n etcd-check create configmap created-after-backup --from-literal=state=lost-on-restore >/dev/null
