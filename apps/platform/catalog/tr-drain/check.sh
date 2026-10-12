NS=tr-drain
cordoned() { eq true "$(kubectl get node cka-sim-worker2 -o jsonpath='{.spec.unschedulable}')"; }
leftovers() {
  kubectl get pods -A --field-selector spec.nodeName=cka-sim-worker2,status.phase!=Succeeded,status.phase!=Failed \
    -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name} {.metadata.ownerReferences[0].kind}{"\n"}{end}' |
    awk '$2 != "DaemonSet"'
}
empty() { [ -z "$(leftovers)" ]; }
drained() { cordoned && empty; }
ready() { eq "$2" "$(kubectl -n $NS get deployment "$1" -o jsonpath='{.status.readyReplicas}')"; }
moved() { drained && ready api 2 && ready cache 1; }
untouched() { moved && eq 1 "$(kubectl -n $NS get deployment api -o jsonpath='{.metadata.generation}')" && eq 1 "$(kubectl -n $NS get deployment cache -o jsonpath='{.metadata.generation}')"; }
cordoned && wait_for 60 empty && wait_for 90 moved

check 2 "Node cka-sim-worker2 is unschedulable" cordoned
check 3 "Only DaemonSet Pods are left on cka-sim-worker2" drained
check 2 "api has 2 and cache has 1 Ready Pods on other nodes" moved
check 1 "The Deployments were not modified" untouched
