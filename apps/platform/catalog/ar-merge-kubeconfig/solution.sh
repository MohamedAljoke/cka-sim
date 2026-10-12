cd /opt/course/ar-merge-kubeconfig
KUBECONFIG=team-a.kubeconfig:team-b.kubeconfig kubectl config view --flatten > merged.kubeconfig
kubectl --kubeconfig merged.kubeconfig config use-context team-b
kubectl --kubeconfig merged.kubeconfig config set-context --current --namespace=ar-merge-b
kubectl --kubeconfig merged.kubeconfig get configmaps
