# crictl ps -a | grep -E 'etcd|apiserver' -> both keep exiting; the API server only because etcd is gone
# crictl logs <etcd-id> 2>&1 | tail -3   -> open /etc/kubernetes/pki/etcd/server.cert: no such file or directory
for id in $(crictl ps -a -q --name '^etcd$'); do crictl logs "$id" 2>&1; done | grep -m1 '"level":"fatal"' > /opt/course/tr-etcd/error.log
sed -i 's|--cert-file=/etc/kubernetes/pki/etcd/server.cert|--cert-file=/etc/kubernetes/pki/etcd/server.crt|' /etc/kubernetes/manifests/etcd.yaml
