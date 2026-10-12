meta() { helm -n ar-helm-reuse get metadata portal 2>/dev/null | awk -F': ' -v key="$1" '$1 == key { print $2 }'; }
values() { helm -n ar-helm-reuse get values portal --all -o json 2>/dev/null; }
upgraded() { [ "$(meta REVISION)" -ge 3 ] 2>/dev/null && eq "deployed 1.1.0" "$(meta STATUS) $(meta VERSION)" && ! helm -n ar-helm-reuse history portal --max 1 | grep -q Rollback; }
all_values() { values | grep -q '"replicaCount":3' && values | grep -q '"type":"NodePort"' && values | grep -q '"pullPolicy":"Always"'; }
running() { eq "3 NodePort Always" "$(kubectl -n ar-helm-reuse get deployment portal-web -o jsonpath='{.status.readyReplicas}') $(kubectl -n ar-helm-reuse get service portal-web -o jsonpath='{.spec.type}') $(kubectl -n ar-helm-reuse get deployment portal-web -o jsonpath='{.spec.template.spec.containers[0].imagePullPolicy}')"; }

check 2 "portal was upgraded on chart 1.1.0, not rolled back" upgraded
check 2 "portal's values set replicaCount 3, service type NodePort and pullPolicy Always" all_values
check 1 "portal runs 3 Pods with pullPolicy Always behind a NodePort Service" wait_for 60 running
