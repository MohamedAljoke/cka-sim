NS=nw-ingress-tls
HOST=secure.nw-ingress-tls.test
secret() { eq kubernetes.io/tls "$(kubectl -n $NS get secret secure-tls -o jsonpath='{.type}')"; }
cert() { served_cert $HOST | grep -q "CN *= *$HOST"; }
https() { cert && eq "secure /" "$(fetch_tls $HOST /)"; }
ingress_tls() { eq "secure-tls $HOST" "$(kubectl -n $NS get ingress secure -o jsonpath='{.spec.tls[0].secretName} {.spec.tls[0].hosts[0]}')"; }

check 1 "Secret secure-tls is a kubernetes.io/tls Secret" secret
check 1 "Ingress secure lists secure-tls for $HOST under tls" ingress_tls
check 3 "https://$HOST answers with the given certificate" https
