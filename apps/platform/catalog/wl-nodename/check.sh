NS=wl-nodename
pod() { kubectl -n $NS get pod "$1" -o jsonpath="$2" 2>/dev/null; }
pinned_ok() { eq "cka-sim-worker2 Running" "$(pod pinned '{.spec.nodeName} {.status.phase}')"; }
unscheduled() { pinned_ok && [ -z "$(kubectl -n $NS get events --field-selector involvedObject.name=pinned,reason=Scheduled -o name)" ]; }
mirror=static-web-cka-sim-worker2
static_ok() { eq "file Running" "$(pod $mirror '{.metadata.annotations.kubernetes\.io/config\.source} {.status.phase}')"; }
manifest() { grep -l 'name: static-web' /etc/kubernetes/manifests/* 2>/dev/null | grep -q .; }
from_file() { static_ok && manifest; }
manifest && wait_for 60 static_ok

check 2 "Pod pinned runs on cka-sim-worker2" pinned_ok
check 1 "No scheduler placed pinned" unscheduled
check 2 "Static Pod static-web runs on cka-sim-worker2" static_ok
check 1 "Its manifest is in the kubelet's static Pod path" from_file
