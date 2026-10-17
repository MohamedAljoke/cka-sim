NS=tr-sidecar-logs
D() { kubectl -n $NS get deployment orders -o jsonpath="$1"; }
containers() { D '{.spec.template.spec.containers[*].name} {.spec.template.spec.initContainers[*].name}'; }
has_streamer() { [[ " $(containers) " == *" log-streamer "* ]]; }
ready() { has_streamer && eq 1 "$(D '{.status.updatedReplicas}')" && eq 1 "$(D '{.status.readyReplicas}')"; }
streams() { ready && kubectl -n $NS logs deploy/orders -c log-streamer --tail=20 | grep -q 'processed order'; }
app_kept() { streams && D '{.spec.template.spec.containers[?(@.name=="app")].command}' | grep -q '/var/log/orders/orders.log'; }
has_streamer && wait_for 90 streams

check 1 "orders has a log-streamer container" has_streamer
check 3 "kubectl logs -c log-streamer shows the app's log lines" streams
check 1 "Container app is unchanged" app_kept
