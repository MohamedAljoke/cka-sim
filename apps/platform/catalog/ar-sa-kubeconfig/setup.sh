fresh_ns ar-sa-kubeconfig
fresh_course
kubectl -n ar-sa-kubeconfig create serviceaccount ci-bot >/dev/null
kubectl -n ar-sa-kubeconfig create role pod-lister --verb=get,list --resource=pods >/dev/null
kubectl -n ar-sa-kubeconfig create rolebinding ci-bot --role=pod-lister --serviceaccount=ar-sa-kubeconfig:ci-bot >/dev/null
