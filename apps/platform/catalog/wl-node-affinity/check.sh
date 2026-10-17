NS=wl-node-affinity
W=cka-sim-worker2
exists() { kubectl -n $NS get "$@" >/dev/null 2>&1; }
field() { kubectl -n $NS get "$1" "$2" -o jsonpath="$3"; }
fast_ok() {
  eq ssd "$(field pod fast '{.spec.nodeSelector.disktype}')" &&
    eq Running "$(field pod fast '{.status.phase}')" && eq $W "$(field pod fast '{.spec.nodeName}')"
}
affinity_keys() { field deployment cache '{.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[*].matchExpressions[*].key}'; }
# Every Pod a ReplicaSet owns: the Deployment's, old ones included until they are gone.
replica_nodes() { kubectl -n $NS get pods -o jsonpath='{range .items[?(@.metadata.ownerReferences[0].kind=="ReplicaSet")]}{.spec.nodeName}{"\n"}{end}' | sort -u; }
cache_ok() {
  [[ " $(affinity_keys) " == *" disktype "* ]] && eq 2 "$(field deployment cache '{.status.readyReplicas}')" && eq $W "$(replica_nodes)"
}

check 1 "cka-sim-worker2 has the label disktype=ssd" eq ssd "$(kubectl get node $W -o jsonpath='{.metadata.labels.disktype}')"
if exists pod fast; then
  check 2 "Pod fast selects disktype=ssd and runs on cka-sim-worker2" wait_for 60 fast_ok
else
  echo "FAIL 2 Pod fast selects disktype=ssd and runs on cka-sim-worker2"
fi
if exists deployment cache; then
  check 3 "Deployment cache requires disktype and runs 2 Ready Pods on cka-sim-worker2" wait_for 60 cache_ok
else
  echo "FAIL 3 Deployment cache requires disktype and runs 2 Ready Pods on cka-sim-worker2"
fi
