NS=nw-coredns-hosts
lookup() { kubectl -n $NS exec client -- nslookup -timeout=2 "$1" 2>/dev/null; }
legacy() { lookup legacy-db.corp | sed -n '/^Name:/,$p' | grep -q '10\.10\.10\.10'; }
others() { legacy && lookup kubernetes.default.svc.cluster.local | grep -q 'Address:.*10\.96\.0\.1'; }
ready() { legacy && eq 2 "$(kubectl -n kube-system get deployment coredns -o jsonpath='{.status.readyReplicas}')"; }
kubectl -n kube-system get configmap coredns -o jsonpath='{.data.Corefile}' | grep -q legacy-db.corp && wait_for 90 legacy

check 3 "legacy-db.corp resolves to 10.10.10.10 from client" legacy
check 1 "Cluster names still resolve" others
check 1 "Both CoreDNS Pods are Ready" ready
