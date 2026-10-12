NS=nw-netpol-ipblock
WEB=$(pod_ip $NS web)
range() { connects $NS near "$WEB" 80 && ! connects $NS far "$WEB" 80; }
excepted() { ! connects $NS banned "$WEB" 80; }
uses_except() { eq "$(cat "$COURSE/cidr") $(cat "$COURSE/except")/32" "$(kubectl -n $NS get networkpolicy web-ipblock -o jsonpath='{.spec.ingress[0].from[0].ipBlock.cidr} {.spec.ingress[0].from[0].ipBlock.except[0]}')"; }

check 2 "near (inside the CIDR) reaches web, far (outside) is blocked" range
check 2 "banned, the excepted IP, is blocked" excepted
check 1 "web-ipblock uses ipBlock with the CIDR and the IP as a /32 except" uses_except
