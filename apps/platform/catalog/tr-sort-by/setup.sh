fresh_ns tr-sort-by
fresh_course
run() {
  kubectl -n tr-sort-by run "$1" --image=busybox:1.36 --command -- sh -c "$2" >/dev/null
  sleep 1.5
}
run zeta 'sleep 1d'
run alpha 'sleep 1d'
run mike 'sleep 30; exit 1'
run kilo 'exit 1'
restarted() { [ "$(kubectl -n tr-sort-by get pod kilo -o jsonpath='{.status.containerStatuses[0].restartCount}')" -ge 1 ] 2>/dev/null; }
wait_for 90 restarted
