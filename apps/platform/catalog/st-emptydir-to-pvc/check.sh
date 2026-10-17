NS=st-emptydir-to-pvc
D() { kubectl -n $NS get deployment notes -o jsonpath="$1"; }
claim() { D '{.spec.template.spec.volumes[?(@.name=="data")].persistentVolumeClaim.claimName}'; }
uses_pvc() { eq notes-data "$(claim)" && eq Bound "$(kubectl -n $NS get pvc notes-data -o jsonpath='{.status.phase}')"; }
ready() { uses_pvc && eq "1 1" "$(D '{.status.updatedReplicas} {.status.readyReplicas}')"; }
starts() { kubectl -n $NS exec deploy/notes -- cat /data/notes.txt 2>/dev/null | sort -u | wc -l; }
survived() { ready && [ "$(starts)" -ge 2 ]; }
uses_pvc && wait_for 90 ready

check 2 "notes keeps /data on the Bound PVC notes-data" uses_pvc
check 1 "notes runs 1 Ready Pod" ready
check 2 "notes.txt holds lines from at least two different Pods" survived
