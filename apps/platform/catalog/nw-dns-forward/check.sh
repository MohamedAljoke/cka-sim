NS=nw-dns-forward
lookup() { kubectl -n $NS exec client -- nslookup -timeout=2 "$1" 2>/dev/null; }
corp() { lookup db.corp.internal | grep -q 'Address:.*10\.10\.10\.50'; }
cluster() { lookup kubernetes.default.svc.cluster.local | grep -q 'Address:.*10\.96\.0\.1'; }
both() { corp && cluster; }
coredns_ready() { corp && eq 2 "$(kubectl -n kube-system get deployment coredns -o jsonpath='{.status.readyReplicas}')"; }
configured() { kubectl -n kube-system get configmap coredns -o jsonpath='{.data.Corefile}' | grep -q 'corp\.internal'; }
configured && wait_for 120 corp

check 3 "db.corp.internal resolves to 10.10.10.50 from Pod client" corp
check 2 "kubernetes.default.svc.cluster.local still resolves" both
check 1 "Both CoreDNS Pods are Ready" coredns_ready
