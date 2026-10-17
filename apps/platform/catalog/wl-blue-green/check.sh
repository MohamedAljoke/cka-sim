NS=wl-blue-green
selector() { kubectl -n $NS get service web -o jsonpath='{.spec.selector.app} {.spec.selector.version}'; }
switched() { eq "web green" "$(selector)"; }
replies() { for _ in 1 2 3 4 5 6; do kubectl -n $NS exec client -- wget -qO- -T 2 http://web 2>/dev/null; done | sort -u | xargs; }
green_only() { switched && eq green "$(replies)"; }
blue_down() { green_only && eq 0 "$(kubectl -n $NS get deployment web-blue -o jsonpath='{.spec.replicas}')"; }
green_kept() { blue_down && eq 2 "$(kubectl -n $NS get deployment web-green -o jsonpath='{.status.readyReplicas}')"; }
switched && wait_for 30 green_only

check 2 "Service web selects version=green" switched
check 2 "Every request to web is answered by green" green_only
check 1 "web-blue still exists with 0 replicas" blue_down
check 1 "web-green still has 2 Ready replicas" green_kept
