NS=nw-gateway-canary
weights() { kubectl -n $NS get httproute shop -o jsonpath='{range .spec.rules[0].backendRefs[*]}{.name}={.weight} {end}' | xargs -n1 | sort | xargs; }
split() {
  local answers canary stable
  answers=$(for _ in $(seq 100); do fetch shop.nw-gateway-canary.test /; echo; done)
  canary=$(echo "$answers" | grep -c '^canary ')
  stable=$(echo "$answers" | grep -c '^stable ')
  [ "$canary" -ge 5 ] && [ "$canary" -le 40 ] && [ $((canary + stable)) -ge 95 ]
}

check 2 "The rule weighs stable 80 and canary 20" eq "canary=20 stable=80" "$(weights)"
check 3 "Of 100 requests, 5 to 40 reach canary and the rest stable" split
