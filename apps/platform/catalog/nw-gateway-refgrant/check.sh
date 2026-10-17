grant() { kubectl -n nw-gateway-refgrant-data get referencegrant allow-web-routes -o jsonpath='{.spec.from[*].kind}/{.spec.from[*].namespace} {.spec.to[*].kind}'; }
resolved() { eq "True ResolvedRefs" "$(condition nw-gateway-refgrant httproute/catalog ResolvedRefs)"; }
served() { eq "catalog /items" "$(fetch catalog.nw-gateway-refgrant.test /items)"; }

check 1 "HTTPRoute catalog's ResolvedRefs condition is True" resolved
check 3 "catalog.nw-gateway-refgrant.test reaches Service catalog" served
check 1 "ReferenceGrant allow-web-routes allows HTTPRoutes from nw-gateway-refgrant to Services" eq "HTTPRoute/nw-gateway-refgrant Service" "$(grant)"
