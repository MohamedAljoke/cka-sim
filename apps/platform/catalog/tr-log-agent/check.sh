NS=tr-log-agent
ds() { kubectl -n $NS get daemonset log-agent -o jsonpath="$1" 2>/dev/null; }
workers() { kubectl get nodes --no-headers -l '!node-role.kubernetes.io/control-plane' | wc -l; }
on_workers() { eq "$(workers) $(workers)" "$(ds '{.status.desiredNumberScheduled} {.status.numberReady}')"; }
mount_ro() { eq "/var/log/pods true" "$(ds '{.spec.template.spec.volumes[?(@.hostPath.path=="/var/log/pods")].name}' | xargs -I{} kubectl -n $NS get daemonset log-agent -o jsonpath='{.spec.template.spec.containers[0].volumeMounts[?(@.name=="{}")].mountPath} {.spec.template.spec.containers[0].volumeMounts[?(@.name=="{}")].readOnly}')"; }
agent() { kubectl -n $NS get pods -l app=log-agent --field-selector spec.nodeName="$(kubectl -n tr-log-agent-apps get pod app -o jsonpath='{.spec.nodeName}')" -o name | head -1; }
ships() { [ -n "$(agent)" ] && kubectl -n $NS logs "$(agent)" --tail=50 | grep -q 'app heartbeat'; }
ds '{.metadata.name}' >/dev/null && wait_for 90 on_workers && wait_for 30 ships

check 2 "log-agent runs one Ready Pod on every worker node" on_workers
check 1 "It mounts the node's /var/log/pods read-only" mount_ro
check 3 "The agent on app's node prints app's log lines" ships
