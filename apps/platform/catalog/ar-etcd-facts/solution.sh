export ETCDCTL_CACERT=/etc/kubernetes/pki/etcd/ca.crt ETCDCTL_CERT=/etc/kubernetes/pki/etcd/server.crt ETCDCTL_KEY=/etc/kubernetes/pki/etcd/server.key
# etcdctl endpoint status -w table -> VERSION column; etcdctl member list -w table -> NAME column
{
  echo "version: $(etcdctl --endpoints=https://127.0.0.1:2379 endpoint status -w json | jq -r '.[0].Status.version')"
  echo "data-dir: $(grep -o -- '--data-dir=[^ ]*' /etc/kubernetes/manifests/etcd.yaml | cut -d= -f2)"
  echo "members: $(etcdctl --endpoints=https://127.0.0.1:2379 member list | cut -d, -f3 | tr -d ' ' | paste -sd,)"
} > /opt/course/ar-etcd-facts/etcd.txt
