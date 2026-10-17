NS=tr-debug
eph() { kubectl -n $NS get pod app -o jsonpath="{.spec.ephemeralContainers[?(@.name==\"debugger\")].$1}"; }
attached() { eq "busybox:1.36 app" "$(eph image) $(eph targetContainerName)"; }
ps_ok() { grep -q '/pause' "$COURSE/ps.txt" 2>/dev/null; }
same_pod() { attached && eq 0 "$(kubectl -n $NS get pod app -o jsonpath='{.status.containerStatuses[0].restartCount}')" && eq 1 "$(kubectl -n $NS get pod app -o jsonpath='{.spec.ephemeralContainers[*].name}' | wc -w)"; }

check 2 "Ephemeral container debugger (busybox:1.36) targets container app" attached
check 2 "ps.txt lists the /pause process of container app" ps_ok
check 1 "Pod app was not restarted" same_pod
