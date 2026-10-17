fresh_ns ar-openssl-user
fresh_course
kubectl -n ar-openssl-user create rolebinding ops-team-view --clusterrole=view --group=ops-team >/dev/null
kubectl -n ar-openssl-user create configmap runbook --from-literal=step1=restart >/dev/null
