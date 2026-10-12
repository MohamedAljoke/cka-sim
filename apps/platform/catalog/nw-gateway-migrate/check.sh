NS=nw-gateway-migrate
HOST=shop.nw-gateway-migrate.test
no_ingress() { ! kubectl -n $NS get ingress shop >/dev/null 2>&1; }
listeners() { kubectl -n $NS get gateway shop-gw -o jsonpath='{range .spec.listeners[*]}{.name}:{.protocol}:{.port}:{.tls.certificateRefs[0].name} {end}' | xargs; }
gw() { eq "http:HTTP:80: https:HTTPS:443:shop-tls" "$(listeners)" && eq "True Programmed" "$(condition $NS gateway/shop-gw Programmed)"; }
http_kept() { no_ingress && eq "store /" "$(fetch $HOST /)" && eq "api /api/x" "$(fetch $HOST /api/x)"; }
https_kept() { no_ingress && served_cert $HOST | grep -q "CN *= *$HOST" && eq "api /api/x" "$(fetch_tls $HOST /api/x)" && eq "store /cart" "$(fetch_tls $HOST /cart)"; }
route() { eq "True Accepted" "$(condition $NS httproute/shop Accepted)"; }

check 2 "Gateway shop-gw has listeners http (80) and https (443, Secret shop-tls) and is Programmed" gw
check 1 "HTTPRoute shop is Accepted" route
check 2 "With Ingress shop deleted, http:// still routes /api to api and / to store" http_kept
check 2 "With Ingress shop deleted, https:// does the same with the shop-tls certificate" https_kept
