NS=st-projected
running() { eq True "$(kubectl -n $NS get pod app -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)"; }
in_pod() { kubectl -n $NS exec app -- cat "/etc/app/$1" 2>/dev/null; }
one_volume() { local v; v=$(kubectl -n $NS get pod app -o jsonpath='{.spec.containers[0].volumeMounts[?(@.mountPath=="/etc/app")].name}' 2>/dev/null); running && [ -n "$v" ] && [ -n "$(kubectl -n $NS get pod app -o jsonpath="{.spec.volumes[?(@.name==\"$v\")].projected.sources}")" ]; }
cm_secret() { running && eq "mode=fast" "$(in_pod settings)" && eq "s3cr3t-token" "$(in_pod token)"; }
labels_ok() { running && in_pod labels | grep -qx 'tier="backend"'; }
aud() {
  local p
  p=$(in_pod vault-token | cut -d. -f2 | tr '_-' '/+')
  while [ $(( ${#p} % 4 )) -ne 0 ]; do p="$p="; done
  echo "$p" | base64 -d 2>/dev/null | jq -c '[.aud, (.exp - .iat)]'
}
vault_ok() { running && eq '[["vault"],3600]' "$(aud)"; }
kubectl -n $NS get pod app >/dev/null 2>&1 && wait_for 60 running

check 1 "/etc/app is one projected volume" one_volume
check 2 "settings and token come from the ConfigMap and the Secret" cm_secret
check 1 "labels holds the Pod's labels" labels_ok
check 2 "vault-token is a ServiceAccount token for audience vault, valid 3600s" vault_ok
