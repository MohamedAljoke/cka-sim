# kubectl get csr dev-anna -o jsonpath='{.spec.request}' | base64 -d | openssl req -noout -subject -> CN = anna, O = developers
# kubectl get clusterrolebinding kubeadm:cluster-admins -o wide                    -> that group is bound to cluster-admin
# kubectl get csr ci-runner -o jsonpath='{.spec.request}' | base64 -d | openssl req -noout -subject -> O = kubeadm:cluster-admins
kubectl certificate approve dev-anna
kubectl certificate deny ci-runner
