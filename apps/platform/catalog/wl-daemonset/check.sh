NS=wl-daemonset
ds() { kubectl -n $NS get daemonset log-agent -o jsonpath="$1"; }
nodes() { kubectl get nodes -o name | wc -l; }
everywhere() { eq "$(nodes) $(nodes)" "$(ds '{.status.desiredNumberScheduled} {.status.numberReady}')"; }
on_cp() { [ -n "$(kubectl -n $NS get pods -l app=log-agent --field-selector spec.nodeName=cka-sim-control-plane -o name)" ]; }
taint_kept() { on_cp && kubectl get node cka-sim-control-plane -o jsonpath='{.spec.taints[*].key}' | grep -q node-role.kubernetes.io/control-plane; }
ds '{.metadata.name}' >/dev/null 2>&1 && wait_for 90 everywhere

check 3 "log-agent runs one Ready Pod on every node" everywhere
check 1 "One of them runs on the control plane" on_cp
check 1 "The control plane keeps its taint" taint_kept
