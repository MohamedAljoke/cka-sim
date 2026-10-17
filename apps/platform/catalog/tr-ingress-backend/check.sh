NS=tr-ingress-backend
HOST=store.tr-ingress-backend.test
services() { eq "api storefront" "$(kubectl -n $NS get services -o jsonpath='{.items[*].metadata.name}')" && eq 80 "$(kubectl -n $NS get service storefront -o jsonpath='{.spec.ports[*].port}')"; }
root() { eq "storefront /" "$(fetch $HOST /)" && services; }
api() { eq "api /api/orders" "$(fetch $HOST /api/orders)" && services; }

check 3 "/ reaches storefront, with the Services unchanged" root
check 3 "/api/orders reaches api, with the Services unchanged" api
