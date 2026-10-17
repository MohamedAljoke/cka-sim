NS=nw-gateway-routes
HOST=app.nw-gateway-routes.test
routes() { eq "api /api/users" "$(fetch $HOST /api/users)" && eq "web /home" "$(fetch $HOST /home)"; }
gw() { eq "traefik http HTTP 80" "$(kubectl -n $NS get gateway web-gw -o jsonpath='{.spec.gatewayClassName} {.spec.listeners[0].name} {.spec.listeners[0].protocol} {.spec.listeners[0].port}')"; }
programmed() { gw && eq "True Programmed" "$(condition $NS gateway/web-gw Programmed)"; }
accepted() { eq "True Accepted" "$(condition $NS httproute/web-route Accepted)"; }

check 1 "Gateway web-gw uses GatewayClass traefik, with listener http on HTTP port 80, and is Programmed" programmed
check 1 "HTTPRoute web-route is Accepted by web-gw" accepted
check 3 "/api/users reaches api, /home reaches web" routes
