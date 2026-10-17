available() { eq 1 "$(kubectl -n kube-system get deployment metrics-server -o jsonpath='{.status.availableReplicas}')"; }
every_node() { [ "$(kubectl top nodes --no-headers 2>/dev/null | wc -l)" -eq 3 ]; }
top_works() { wait_for 90 every_node; }

check 2 "Deployment metrics-server in kube-system is available" available
check 3 "kubectl top nodes shows every node" top_works
