NS=st-readonly-share
running() { eq True "$(kubectl -n $NS get pod "$1" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)"; }
mount() { kubectl -n $NS get pod "$1" -o jsonpath="{.spec.containers[0].volumeMounts[?(@.mountPath==\"$2\")].$3}" 2>/dev/null; }
claim() { kubectl -n $NS get pod "$1" -o jsonpath="{.spec.volumes[?(@.persistentVolumeClaim.claimName==\"shared\")].name}" 2>/dev/null; }
writer_ok() { running writer && eq writer-logs "$(mount writer /logs subPath)" && [ -n "$(claim writer)" ]; }
reads() { running auditor && kubectl -n $NS exec auditor -- grep -q . /audit/writer-logs/app.log; }
read_only() { reads && eq true "$(mount auditor /audit readOnly)" && ! kubectl -n $NS exec auditor -- touch /audit/probe 2>/dev/null; }
kubectl -n $NS get pod writer auditor >/dev/null 2>&1 && wait_for 90 reads

check 2 "writer mounts sub-folder writer-logs of PVC shared at /logs" writer_ok
check 2 "auditor reads /audit/writer-logs/app.log" reads
check 2 "auditor's mount is read-only and refuses writes" read_only
