replicas() { kubectl -n wl-scale get deployment web -o jsonpath='{.spec.replicas}'; }
ready() { kubectl -n wl-scale get deployment web -o jsonpath='{.status.readyReplicas}'; }
four_ready() { eq 4 "$(ready)"; }

check 2 "Deployment web wants 4 replicas" eq 4 "$(replicas)"

# Only after the scale: 1 Ready pod mustn't earn points for doing nothing.
if eq 4 "$(replicas)"; then
  check 2 "4 Pods of web are Ready" wait_for 60 four_ready
else
  echo "FAIL 2 4 Pods of web are Ready"
fi
