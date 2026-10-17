NS=nw-ingress-rules
HOST=shop.nw-ingress-rules.test
api() { eq "api /api/orders" "$(fetch $HOST /api/orders)"; }
store() { eq "store /cart" "$(fetch $HOST /cart)"; }
class() { eq nginx "$(kubectl -n $NS get ingress shop -o jsonpath='{.spec.ingressClassName}')"; }
saved() { grep -qx "api /api/orders" "$COURSE/curl.txt"; }

check 2 "/api/orders on $HOST reaches api" api
check 2 "/cart on $HOST reaches store" store
check 1 "Ingress shop uses IngressClass nginx" class
check 1 "curl.txt holds api's answer for /api/orders" saved
