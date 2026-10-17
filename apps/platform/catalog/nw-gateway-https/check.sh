NS=nw-gateway-https
HOST=secure.nw-gateway-https.test
https() { served_cert $HOST | grep -q "CN *= *$HOST" && eq "app /x" "$(fetch_tls $HOST /x)"; }
redirect() { eq "301 https://$HOST/x" "$(fetch $HOST /x -o /dev/null -w '%{http_code} %{redirect_url}')"; }
listener() { kubectl -n $NS get gateway secure-gw -o jsonpath='{.spec.listeners[?(@.name=="https")].protocol}:{.spec.listeners[?(@.name=="https")].port}:{.spec.listeners[?(@.name=="https")].tls.certificateRefs[0].name}'; }
sections() { kubectl -n $NS get httproute "$1" -o jsonpath='{.spec.parentRefs[*].sectionName}'; }
attached() { eq https "$(sections app)" && eq http "$(sections redirect)"; }

check 1 "secure-gw has listener https on HTTPS port 443 with Secret secure-tls" eq "HTTPS:443:secure-tls" "$(listener)"
check 3 "https://$HOST/x reaches app with the secure-tls certificate" https
check 2 "http://$HOST/x answers 301 to https://$HOST/x" redirect
check 1 "Route app attaches to listener https, route redirect to listener http" attached
