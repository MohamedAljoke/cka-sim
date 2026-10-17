NS=wl-psa
label() { kubectl get namespace $NS -o jsonpath="{.metadata.labels.pod-security\.kubernetes\.io/$1}"; }
enforced() { eq restricted "$(label enforce)" && { [ -z "$(label enforce-version)" ] || eq latest "$(label enforce-version)"; }; }
D() { kubectl -n $NS get deployment web -o jsonpath="$1"; }
ready() { eq "2 2" "$(D '{.status.updatedReplicas} {.status.readyReplicas}')" && [ -z "$(D '{.status.unavailableReplicas}')" ]; }
admitted() {
  enforced && ready || return 1
  local pod
  for pod in $(kubectl -n $NS get pods -l app=web -o name); do
    kubectl -n $NS get "$pod" -o json | jq 'del(.metadata.resourceVersion,.metadata.uid,.metadata.creationTimestamp,.metadata.ownerReferences,.status) | .metadata.name += "-probe"' |
      kubectl create --dry-run=server -f - >/dev/null 2>&1 || return 1
  done
}
kept() { admitted && eq busybox:1.36 "$(D '{.spec.template.spec.containers[0].image}')" && eq '["sleep","1d"]' "$(D '{.spec.template.spec.containers[0].command}')"; }
enforced && wait_for 90 admitted

check 2 "Namespace wl-psa enforces restricted" enforced
check 3 "web's Pods pass the restricted standard and 2 are Ready" admitted
check 1 "web keeps its image and command" kept
