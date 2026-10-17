NS=nw-gateway-header
HOST=app.nw-gateway-header.test
v2() { eq "v2 /" "$(fetch $HOST / -H 'x-version: v2')"; }
others() { v2 && eq "v1 /" "$(fetch $HOST /)" && eq "v1 /" "$(fetch $HOST / -H 'x-version: v3')"; }
match() { kubectl -n $NS get httproute app -o jsonpath='{.spec.rules[*].matches[*].headers[*].name}' | grep -qix x-version; }

check 3 "Requests with x-version: v2 reach v2" v2
check 1 "Requests without it, or with another value, still reach v1" others
check 1 "HTTPRoute app matches on header x-version" match
