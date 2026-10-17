pod() { kubectl -n wl-pod get pod web -o jsonpath="$1" 2>/dev/null; }
running() { eq Running "$(pod '{.status.phase}')"; }

check 1 "Pod web uses image nginx:1.27" eq nginx:1.27 "$(pod '{.spec.containers[0].image}')"
if pod '{.metadata.name}' >/dev/null; then
  check 1 "Pod web is Running" wait_for 60 running
else
  echo "FAIL 1 Pod web is Running"
fi
