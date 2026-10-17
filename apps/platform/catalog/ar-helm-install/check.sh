meta() { helm -n ar-helm-install get metadata shop ${2:+--revision "$2"} 2>/dev/null | awk -F': ' -v key="$1" '$1 == key { print $2 }'; }
values() { helm -n ar-helm-install get values shop ${1:+--revision "$1"} 2>/dev/null; }
installed() { eq "1.0.0" "$(meta VERSION 1)" && values 1 | grep -qx 'replicaCount: 3'; }
upgraded() { eq "1.1.0" "$(meta VERSION 2)" && values 2 | grep -qx 'replicaCount: 3'; }
rolled_back() { eq "3 deployed 1.0.0" "$(meta REVISION) $(meta STATUS) $(meta VERSION)"; }
ready() { eq "3 nginx:1.26-alpine" "$(kubectl -n ar-helm-install get deployment shop-web -o jsonpath='{.status.readyReplicas} {.spec.template.spec.containers[0].image}' 2>/dev/null)"; }

check 1 "revision 1 installed web 1.0.0 with replicaCount 3" installed
check 2 "revision 2 upgraded to web 1.1.0 and kept replicaCount 3" upgraded
check 2 "revision 3 rolled back to web 1.0.0 and is deployed" rolled_back
check 1 "Deployment shop-web runs 3 Pods of nginx:1.26-alpine" wait_for 60 ready
