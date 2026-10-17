answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/pv.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
volume() { kubectl -n st-reclaim get pvc orders -o jsonpath='{.spec.volumeName}'; }
policy() { kubectl get pv "$(volume)" -o jsonpath='{.spec.persistentVolumeReclaimPolicy}'; }

check 1 "pv.txt names the PV of orders" eq "$(volume)" "$(answer pv)"
check 1 "pv.txt gives its original policy, Delete" eq Delete "$(answer policy)"
check 3 "The PV now has the reclaim policy Retain" eq Retain "$(policy)"
