meta() { helm -n ar-helm-stuck get metadata billing 2>/dev/null | awk -F': ' -v key="$1" '$1 == key { print $2 }'; }
deployed() { eq "deployed 1.0.0" "$(meta STATUS) $(meta VERSION)"; }
kept_history() { [ "$(meta REVISION)" -ge 3 ] 2>/dev/null; }
running() { eq "1 nginx:1.26-alpine" "$(kubectl -n ar-helm-stuck get deployment billing-web -o jsonpath='{.status.readyReplicas} {.spec.template.spec.containers[0].image}' 2>/dev/null)"; }

check 2 "billing is deployed on web 1.0.0" deployed
check 1 "billing was rolled back, not reinstalled" kept_history
check 2 "Deployment billing-web runs nginx:1.26-alpine" wait_for 60 running
