NS=st-writable-layer
P() { kubectl -n $NS get pod scratch -o jsonpath="$1" 2>/dev/null; }
mounted() { local v; v=$(P '{.spec.containers[0].volumeMounts[?(@.mountPath=="/cache")].name}'); [ -n "$v" ] && [[ " $(P '{.spec.volumes[?(@.emptyDir)].name}') " == *" $v "* ]]; }
restarted() { mounted && [ "$(P '{.status.containerStatuses[0].restartCount}')" -ge 1 ] 2>/dev/null && eq true "$(P '{.status.containerStatuses[0].ready}')"; }
kept() {
  restarted || return 1
  local written started
  written=$(kubectl -n $NS exec scratch -- cat /cache/marker 2>/dev/null)
  started=$(date -d "$(P '{.status.containerStatuses[0].state.running.startedAt}')" +%s)
  [ -n "$written" ] && [ "$written" -lt "$started" ]
}
mounted && wait_for 60 restarted

check 2 "scratch mounts an emptyDir at /cache" mounted
check 1 "Its container was restarted at least once" restarted
check 2 "/cache/marker survived the restart" kept
