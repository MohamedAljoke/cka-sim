attached() { eq "True Accepted" "$(condition nw-gateway-allowed-team httproute/team Accepted)" && eq "app /" "$(fetch team.nw-gateway-allowed.test /)"; }
others_out() { attached && ! eq "True Accepted" "$(condition nw-gateway-allowed-other httproute/team Accepted)" && ! fetch other.nw-gateway-allowed.test / | grep -q '^app '; }
selector() { kubectl -n nw-gateway-allowed get gateway shared-gw -o jsonpath='{.spec.listeners[0].allowedRoutes.namespaces.from} {.spec.listeners[0].allowedRoutes.namespaces.selector.matchLabels.shared-gateway-access}'; }
route_kept() { eq "shared-gw nw-gateway-allowed" "$(kubectl -n nw-gateway-allowed-team get httproute team -o jsonpath='{.spec.parentRefs[0].name} {.spec.parentRefs[0].namespace}')"; }

check 3 "HTTPRoute team is Accepted and team.nw-gateway-allowed.test answers" attached
check 1 "The same route in unlabelled Namespace nw-gateway-allowed-other stays out" others_out
check 1 "Listener http selects Namespaces with shared-gateway-access=true" eq "Selector true" "$(selector)"
