# The flags for etcdctl come from the etcd static Pod:
#   grep -E 'cert|key|trusted-ca|listen-client' /etc/kubernetes/manifests/etcd.yaml
set -e
etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key \
  snapshot save "$COURSE/etcd-snapshot.db"

dir=/var/lib/etcd-restore-$(date +%s)
# Stop the API server so nothing writes to etcd or serves cached state during the swap.
mv /etc/kubernetes/manifests/kube-apiserver.yaml /root/
# --bump-revision/--mark-compacted make every watcher notice that history changed and re-list.
etcdutl snapshot restore "$COURSE/etcd-backup-old.db" --data-dir "$dir" \
  --bump-revision 1000000000 --mark-compacted
old=$(crictl ps -q --name '^etcd$')
sed -i "s|^\(      path: \)/var/lib/etcd.*\$|\1$dir|" /etc/kubernetes/manifests/etcd.yaml
# An API server started before the new etcd would cache the pre-restore objects.
until new=$(crictl ps -q --name '^etcd$') && [ -n "$new" ] && [ "$new" != "$old" ]; do sleep 2; done
mv /root/kube-apiserver.yaml /etc/kubernetes/manifests/
