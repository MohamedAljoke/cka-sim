pvc() { kubectl -n st-volumename get pvc orders -o jsonpath="$1" 2>/dev/null; }
bound_b() { eq "Bound pv-orders-b" "$(pvc '{.status.phase} {.spec.volumeName}')"; }
a_free() { bound_b && eq Available "$(kubectl get pv pv-orders-a -o jsonpath='{.status.phase}')"; }
pvc '{.metadata.name}' >/dev/null && wait_for 30 bound_b

check 3 "PVC orders is Bound to pv-orders-b" bound_b
check 1 "pv-orders-a is still Available" a_free
