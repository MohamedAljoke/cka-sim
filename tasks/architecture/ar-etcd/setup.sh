ETCDCTL="etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key"
fresh_ns etcd-check
k -n etcd-check create configmap marker --from-literal=state=from-backup >/dev/null
on_host "rm -rf $(course_dir) && mkdir -p $(course_dir) && $ETCDCTL snapshot save $(course_dir)/etcd-backup-old.db >/dev/null"
k -n etcd-check delete configmap marker >/dev/null
k -n etcd-check create configmap created-after-backup --from-literal=state=lost-on-restore >/dev/null
