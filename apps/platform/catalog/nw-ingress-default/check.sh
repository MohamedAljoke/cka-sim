NS=nw-ingress-default
backend() { kubectl -n $NS get ingress site -o jsonpath='{.spec.defaultBackend.service.name}:{.spec.defaultBackend.service.port.number}'; }
caught() { eq "fallback /blog" "$(fetch www.nw-ingress-default.test /blog)" && eq "site /shop/cart" "$(fetch www.nw-ingress-default.test /shop/cart)"; }

check 3 "/blog reaches fallback, while /shop still reaches site" caught
check 2 "Ingress site has defaultBackend fallback:80" eq fallback:80 "$(backend)"
