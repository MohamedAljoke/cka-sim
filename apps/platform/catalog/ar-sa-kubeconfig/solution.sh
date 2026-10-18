KC=/opt/course/ar-sa-kubeconfig/ci-bot.kubeconfig
kubectl config view --raw -o jsonpath='{.clusters[?(@.name=="cka-sim")].cluster.certificate-authority-data}' | base64 -d > /tmp/ca.crt
kubectl --kubeconfig $KC config set-cluster cka-sim --server=https://cka-sim-control-plane:6443 --certificate-authority=/tmp/ca.crt --embed-certs
kubectl --kubeconfig $KC config set-credentials ci-bot --token="$(kubectl -n ar-sa-kubeconfig create token ci-bot --duration=1h)"
kubectl --kubeconfig $KC config set-context ci-bot --cluster=cka-sim --user=ci-bot --namespace=ar-sa-kubeconfig
kubectl --kubeconfig $KC config use-context ci-bot
kubectl --kubeconfig $KC get pods
