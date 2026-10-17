NS=tr-crashloop
template() { kubectl -n $NS get deployment api -o jsonpath="{.spec.template.spec.$1}"; }
has_key() { [ -n "$(kubectl -n $NS get configmap api-config -o jsonpath='{.data.config\.yaml}')" ]; }
ready() { eq 1 "$(kubectl -n $NS get deployment api -o jsonpath='{.status.readyReplicas}')"; }
untouched() {
  eq "busybox:1.36" "$(template 'containers[*].image')" &&
    eq '{"defaultMode":420,"name":"api-config"}' "$(template 'volumes[*].configMap')" &&
    eq /etc/api "$(template 'containers[*].volumeMounts[*].mountPath')" &&
    template 'containers[0].command[2]' | grep -qF '[ -f /etc/api/config.yaml ]'
}

check 1 "error.txt holds the FATAL line" grep -qF 'FATAL: /etc/api/config.yaml not found' "$COURSE/error.txt"
check 2 "ConfigMap api-config has the key config.yaml" has_key
if has_key || ready; then
  check 3 "Deployment api runs 1 Ready Pod" wait_for 90 ready
else
  echo "FAIL 3 Deployment api runs 1 Ready Pod"
fi
if ready; then
  check 1 "The containers and volumes of api were not changed" untouched
else
  echo "FAIL 1 The containers and volumes of api were not changed"
fi
