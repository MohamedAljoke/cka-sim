NS=nw-fixed-clusterip
fixed() { eq "10.96.200.200 80" "$(kubectl -n $NS get service cache -o jsonpath='{.spec.clusterIP} {.spec.ports[0].port}' 2>/dev/null)"; }
serves() { fixed && kubectl -n $NS exec client -- wget -qO- -T 3 http://10.96.200.200 2>/dev/null | grep -q 'Welcome to nginx'; }
rejected() { grep -qiE 'not in the valid range|does not match the current range' "$COURSE/error.txt" 2>/dev/null && ! kubectl -n $NS get service cache-outside >/dev/null 2>&1; }
fixed && wait_for 20 serves

check 2 "Service cache has ClusterIP 10.96.200.200" fixed
check 2 "http://10.96.200.200 serves nginx" serves
check 1 "error.txt holds the out-of-range error and cache-outside doesn't exist" rejected
