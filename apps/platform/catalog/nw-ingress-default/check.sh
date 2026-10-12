NS=nw-ingress-default
backend() { kubectl -n $NS get ingress site -o jsonpath='{.spec.defaultBackend.service.name}:{.spec.defaultBackend.service.port.number}'; }
caught() { eq "fallback /x" "$(fetch other.nw-ingress-default.test /x)" && eq "site /" "$(fetch www.nw-ingress-default.test /)"; }

check 3 "Another host reaches fallback, while www.nw-ingress-default.test still reaches site" caught
check 2 "Ingress site has defaultBackend fallback:80" eq fallback:80 "$(backend)"
