manifest=/etc/kubernetes/manifests/kube-controller-manager.yaml
NS=tr-controller-manager
flag_ok() { grep -q '^    - --kubeconfig=/etc/kubernetes/controller-manager.conf$' "$manifest"; }
cm_ready() { crictl ps -q --name '^kube-controller-manager$' | grep -q . && eq True "$(kubectl -n kube-system get pod kube-controller-manager-cka-sim-control-plane -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
ready() { eq 2 "$(kubectl -n $NS get deployment web -o jsonpath='{.status.readyReplicas}')"; }
created_before_fix() {
  local started
  started=$(crictl inspect "$(crictl ps -q --name '^kube-controller-manager$' | head -1)" | jq -r '.status.startedAt')
  [ "$(date -d "$(kubectl -n $NS get deployment web -o jsonpath='{.metadata.creationTimestamp}')" +%s)" -lt "$(date -d "$started" +%s)" ]
}
untouched() { ready && created_before_fix && eq 1 "$(kubectl -n $NS get deployment web -o jsonpath='{.metadata.generation}')"; }
if flag_ok; then
  wait_for 90 cm_ready
  wait_for 90 ready
fi

check 2 "kube-controller-manager runs and is Ready" cm_ready
check 3 "Deployment web has 2 Ready Pods" ready
check 1 "Deployment web was not modified or recreated" untouched
