NS=st-secret-items
running() { eq True "$(kubectl -n $NS get pod web -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)"; }
in_pod() { kubectl -n $NS exec web -- "$@" 2>/dev/null; }
only_key() { running && eq server.key "$(in_pod ls /etc/ssl/private | xargs)" && eq '-----PRIVATE KEY-----' "$(in_pod cat /etc/ssl/private/server.key)"; }
mode_ok() { running && eq 400 "$(in_pod stat -L -c %a /etc/ssl/private/server.key)"; }
kubectl -n $NS get pod web >/dev/null 2>&1 && wait_for 60 running

check 3 "/etc/ssl/private holds only server.key with the content of tls.key" only_key
check 2 "server.key has mode 0400" mode_ok
