exists() { kubectl -n galaxy get "$@" >/dev/null 2>&1; }
probe_running() { eq Running "$(kubectl -n galaxy get pod probe -o jsonpath='{.status.phase}')"; }
web_ready() { eq 2 "$(kubectl -n galaxy get deployment web -o jsonpath='{.status.readyReplicas}')"; }
written() { tr -d '[:space:]' < "$COURSE/hermes.txt" 2>/dev/null; }
found() {
  local ns
  ns=$(kubectl get pods -A --field-selector metadata.name=hermes -o jsonpath='{.items[0].metadata.namespace}')
  [ -n "$ns" ] && eq "$ns" "$(written)"
}

check 1 "Namespace galaxy has the label env=training" eq training "$(kubectl get namespace galaxy -o jsonpath='{.metadata.labels.env}' 2>/dev/null)"
if exists pod probe; then
  check 1 "Pod probe is Running in galaxy" wait_for 60 probe_running
else
  echo "FAIL 1 Pod probe is Running in galaxy"
fi
if exists deployment web; then
  check 2 "Deployment web has 2 Ready Pods in galaxy" wait_for 60 web_ready
else
  echo "FAIL 2 Deployment web has 2 Ready Pods in galaxy"
fi
check 2 "hermes.txt names the Namespace of Pod hermes" found
