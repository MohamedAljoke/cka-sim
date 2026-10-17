answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/route.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
cidr() { kubectl get node cka-sim-worker2 -o jsonpath='{.spec.podCIDR}'; }
route() { ip route get "$(cidr | cut -d/ -f1 | awk -F. '{print $1"."$2"."$3"."$4+1}')"; }
field() { route | grep -o "$1 [^ ]*" | head -1 | cut -d' ' -f2; }

check 1 "podCIDR is cka-sim-worker2's Pod CIDR" eq "$(cidr)" "$(answer podCIDR)"
check 2 "via is the next hop for it" eq "$(field via)" "$(answer via)"
check 1 "dev is the interface it leaves through" eq "$(field dev)" "$(answer dev)"
