NS=ar-no-automount
sa_ok() { eq false "$(kubectl -n $NS get serviceaccount batch-runner -o jsonpath='{.automountServiceAccountToken}' 2>/dev/null)"; }
P() { kubectl -n $NS get pod runner -o jsonpath="$1" 2>/dev/null; }
pod_ok() { eq "batch-runner false" "$(P '{.spec.serviceAccountName} {.spec.automountServiceAccountToken}')"; }
running() { eq True "$(P '{.status.conditions[?(@.type=="Ready")].status}')"; }
no_token() { pod_ok && running && ! kubectl -n $NS exec runner -- ls /var/run/secrets/kubernetes.io/serviceaccount >/dev/null 2>&1; }
pod_ok && wait_for 60 running

check 2 "ServiceAccount batch-runner doesn't automount its token" sa_ok
check 1 "Pod runner uses batch-runner and opts out itself" pod_ok
check 2 "runner has no token directory" no_token
