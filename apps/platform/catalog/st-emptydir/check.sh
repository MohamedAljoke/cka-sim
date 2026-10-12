NS=st-emptydir
running() { eq "Running" "$(kubectl -n $NS get pod "$1" -o jsonpath='{.status.phase}' 2>/dev/null)"; }
sees() { running "$1" && [ -n "$(kubectl -n $NS exec "$1" -c reader -- cat /data/out.txt 2>/dev/null)" ]; }
shared() { sees share; }
vol() { kubectl -n $NS get pod share-mem -o jsonpath="{.spec.volumes[?(@.emptyDir)].emptyDir.$1}"; }
in_memory() { sees share-mem && eq "Memory 64Mi" "$(vol medium) $(vol sizeLimit)" && kubectl -n $NS exec share-mem -c reader -- grep -q ' /data tmpfs ' /proc/mounts; }
for p in share share-mem; do kubectl -n $NS get pod $p >/dev/null 2>&1 && wait_for 60 sees $p; done

check 3 "In Pod share, reader sees the file writer writes" shared
check 2 "Pod share-mem shares a memory-backed emptyDir limited to 64Mi" in_memory
