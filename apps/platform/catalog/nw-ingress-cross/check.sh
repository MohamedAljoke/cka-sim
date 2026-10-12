HOST=portal.nw-ingress-cross.test
routed() { eq "billing /billing/invoices" "$(fetch $HOST /billing/invoices)" && eq "portal /" "$(fetch $HOST /)"; }
portal_paths() { kubectl -n nw-ingress-cross get ingress portal -o jsonpath='{.spec.rules[*].http.paths[*].backend.service.name}'; }
billing_ingress() { eq nginx "$(kubectl -n nw-ingress-cross-billing get ingress billing -o jsonpath='{.spec.ingressClassName}')"; }
services_stay() { ! kubectl -n nw-ingress-cross get service billing && ! kubectl -n nw-ingress-cross-billing get service portal; }
clean() { eq portal "$(portal_paths)" && services_stay; }

check 3 "/billing/invoices reaches billing, and / still reaches portal" routed
check 1 "Ingress billing exists in nw-ingress-cross-billing with class nginx" billing_ingress
check 1 "Ingress portal only routes to portal, and no Service was copied" clean
