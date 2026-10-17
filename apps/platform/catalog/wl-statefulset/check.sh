NS=wl-statefulset
headless() { eq "None app=db" "$(kubectl -n $NS get service db -o jsonpath='{.spec.clusterIP} app={.spec.selector.app}')"; }
sts() { kubectl -n $NS get statefulset db -o jsonpath="$1"; }
ready() { eq "3 db" "$(sts '{.status.readyReplicas} {.spec.serviceName}')"; }
pvcs() { kubectl -n $NS get pvc data-db-0 data-db-1 data-db-2 -o jsonpath='{range .items[*]}{.status.phase}/{.spec.resources.requests.storage}/{.spec.storageClassName} {end}'; }
bound() { eq "Bound/100Mi/standard Bound/100Mi/standard Bound/100Mi/standard" "$(pvcs | xargs)"; }
dns() { ready && kubectl -n $NS exec db-0 -- nslookup db-1.db.$NS.svc.cluster.local >/dev/null; }
sts '{.metadata.name}' >/dev/null 2>&1 && wait_for 180 ready

check 1 "Service db is headless and selects app=db" headless
check 2 "StatefulSet db has 3 Ready Pods and uses Service db" ready
check 2 "PVCs data-db-0..2 are Bound, 100Mi, class standard" bound
check 1 "db-1.db.wl-statefulset.svc.cluster.local resolves from db-0" dns
