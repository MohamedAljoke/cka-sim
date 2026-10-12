fresh_course
manifest=/etc/kubernetes/manifests/etcd.yaml
sed -i 's|^    - --cert-file=/etc/kubernetes/pki/etcd/server.cert$|    - --cert-file=/etc/kubernetes/pki/etcd/server.crt|' "$manifest"
sed -i 's|^    - --cert-file=/etc/kubernetes/pki/etcd/server.crt$|    - --cert-file=/etc/kubernetes/pki/etcd/server.cert|' "$manifest"

failed() { for id in $(crictl ps -a -q --name '^etcd$'); do crictl logs "$id" 2>&1; done | grep -q 'server\.cert: no such file'; }
wait_for 180 failed
