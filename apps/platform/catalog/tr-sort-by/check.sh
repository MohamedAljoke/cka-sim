listing() { timeout 30 bash "$COURSE/by-age.sh" 2>/dev/null; }
in_order() { eq "zeta alpha mike kilo" "$(listing | grep -owE 'zeta|alpha|mike|kilo' | xargs)"; }
all_namespaces() { listing | grep -q kube-apiserver-cka-sim-control-plane; }
by_age() { in_order && all_namespaces; }
answer() { tr -d '[:space:]' < "$COURSE/restarts.txt" 2>/dev/null; }

check 3 "by-age.sh lists the Pods of all Namespaces, oldest first" by_age
check 2 "restarts.txt names kilo" eq kilo "$(answer)"
