crt="$COURSE/ops.crt"
K() { kubectl --kubeconfig "$COURSE/ops.kubeconfig" --request-timeout=5s "$@"; }
signed() { openssl verify -CAfile /etc/kubernetes/pki/ca.crt "$crt" >/dev/null 2>&1; }
subject_ok() { signed && openssl x509 -in "$crt" -noout -subject | grep -q 'O *= *ops-team' && openssl x509 -in "$crt" -noout -subject | grep -q 'CN *= *ops'; }
days() { echo $(( ( $(date -d "$(openssl x509 -in "$crt" -noout -enddate | cut -d= -f2)" +%s) - $(date -d "$(openssl x509 -in "$crt" -noout -startdate | cut -d= -f2)" +%s) ) / 86400 )); }
thirty() { signed && [ "$(days)" -ge 29 ] && [ "$(days)" -le 31 ]; }
identity() { K auth whoami -o jsonpath='{.status.userInfo.username} {.status.userInfo.groups}' 2>/dev/null; }
logs_in() { [[ "$(identity)" == "ops "*ops-team* ]]; }
reads() { logs_in && K get configmap runbook >/dev/null && eq ar-openssl-user "$(K config view -o jsonpath='{.contexts[?(@.name=="ops")].context.namespace}')"; }

check 2 "ops.crt is signed by the cluster CA for CN=ops, O=ops-team" subject_ok
check 1 "It is valid for 30 days" thirty
check 2 "ops.kubeconfig logs in as ops in group ops-team" logs_in
check 1 "Context ops defaults to ar-openssl-user, where ops can read" reads
