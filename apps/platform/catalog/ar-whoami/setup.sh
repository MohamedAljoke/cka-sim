fresh_ns ar-whoami
fresh_course
kubectl -n ar-whoami create serviceaccount viewer >/dev/null
kubectl -n ar-whoami create role config-reader --verb=get,list --resource=configmaps >/dev/null
kubectl -n ar-whoami create rolebinding viewer --role=config-reader --serviceaccount=ar-whoami:viewer >/dev/null
