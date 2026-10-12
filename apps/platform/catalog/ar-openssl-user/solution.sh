cd /opt/course/ar-openssl-user
openssl genrsa -out ops.key 2048
openssl req -new -key ops.key -subj "/CN=ops/O=ops-team" -out ops.csr
openssl x509 -req -in ops.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 30 -out ops.crt
KC=ops.kubeconfig
kubectl --kubeconfig $KC config set-cluster cka-sim --server=https://cka-sim-control-plane:6443 --certificate-authority=/etc/kubernetes/pki/ca.crt --embed-certs
kubectl --kubeconfig $KC config set-credentials ops --client-certificate=ops.crt --client-key=ops.key --embed-certs
kubectl --kubeconfig $KC config set-context ops --cluster=cka-sim --user=ops --namespace=ar-openssl-user
kubectl --kubeconfig $KC config use-context ops
kubectl --kubeconfig $KC get configmaps
