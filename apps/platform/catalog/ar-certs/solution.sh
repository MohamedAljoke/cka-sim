kubeadm certs check-expiration | awk '$1=="ca"{print $2, $3, $4}' > /opt/course/ar-certs/ca-expiry
kubeadm certs renew apiserver
crictl stop "$(crictl ps -q --name '^kube-apiserver$')"
until kubectl get --raw /readyz >/dev/null 2>&1; do sleep 2; done
