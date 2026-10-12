answer() { tr -s ' \t\r' ' ' < "$COURSE/condition.txt" 2>/dev/null | sed 's/^ //; s/ $//' | head -1; }
fixed() { eq "inventory /stock" "$(fetch inventory.nw-gateway-status.test /stock)" && eq "True ResolvedRefs" "$(condition nw-gateway-status httproute/inventory ResolvedRefs)"; }

check 2 "condition.txt names ResolvedRefs BackendNotFound" eq "ResolvedRefs BackendNotFound" "$(answer)"
check 3 "The route reaches Service inventory, and ResolvedRefs is True" fixed
