# The flags for etcdctl come from the etcd static Pod:
#   grep -E 'cert|key|trusted-ca|listen-client' /etc/kubernetes/manifests/etcd.yaml
ETCDCTL="etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key"
on_host "$ETCDCTL snapshot save $(course_dir)/etcd-snapshot.db"

dir=/var/lib/etcd-restore-$(date +%s)
on_host "set -e
  # 1. Stop the API server so nothing writes to etcd or serves cached state during the swap.
  mv /etc/kubernetes/manifests/kube-apiserver.yaml /root/
  # 2. Restore into a NEW data dir. --bump-revision/--mark-compacted make every watcher
  #    (controllers, the API server's watch cache) notice that history changed and re-list.
  etcdutl snapshot restore $(course_dir)/etcd-backup-old.db --data-dir $dir \
    --bump-revision 1000000000 --mark-compacted
  # 3. Point the etcd static Pod's hostPath volume at it; the kubelet restarts etcd.
  old=\$(crictl ps -q --name '^etcd\$')
  sed -i 's|^\(      path: \)/var/lib/etcd.*\$|\1$dir|' /etc/kubernetes/manifests/etcd.yaml
  # 4. Wait until a NEW etcd container runs — an API server started earlier would talk to the
  #    old etcd and cache pre-restore objects.
  until new=\$(crictl ps -q --name '^etcd\$') && [ -n \"\$new\" ] && [ \"\$new\" != \"\$old\" ]; do sleep 2; done
  # 5. Bring the API server back.
  mv /root/kube-apiserver.yaml /etc/kubernetes/manifests/"
