cond() { kubectl get csr "$1" -o jsonpath="{.status.conditions[?(@.type==\"$2\")].status}" 2>/dev/null; }
anna() { eq True "$(cond dev-anna Approved)"; }
runner() { eq True "$(cond ci-runner Denied)" && [ -z "$(cond ci-runner Approved)" ]; }

check 2 "dev-anna is approved" anna
check 3 "ci-runner, which asks for kubeadm:cluster-admins, is denied" runner
