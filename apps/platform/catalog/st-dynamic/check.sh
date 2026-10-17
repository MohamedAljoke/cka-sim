pv() { tr -d ' \r\n' < "$COURSE/pv.txt" 2>/dev/null; }
named() { [[ "$(pv)" =~ ^pvc-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]]; }
policy() { named && eq Delete "$(tr -d ' \r\n' < "$COURSE/policy.txt" 2>/dev/null)"; }
gone() { named && ! kubectl get pv "$(pv)" && ! kubectl -n st-dynamic get pvc data && ! kubectl -n st-dynamic get pod writer; }
named && wait_for 60 gone

check 2 "pv.txt holds the name of a dynamically provisioned PV" named
check 1 "policy.txt says Delete" policy
check 2 "The Pod, the PVC and its PV are all gone" gone
