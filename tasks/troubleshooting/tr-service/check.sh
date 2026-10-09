endpoints() { [ -n "$(k -n shop get endpointslices -l kubernetes.io/service-name=web -o jsonpath='{.items[*].endpoints[*].addresses[*]}')" ]; }
serves() { [[ "$(k -n shop exec deploy/web -- wget -qO- -T 3 http://web.shop 2>/dev/null)" == *"Welcome to nginx"* ]]; }
wait_for 20 serves
check 2 "Service web has endpoints" endpoints
check 4 "http://web.shop returns the nginx page" serves
untouched() { serves && eq "$(recall generation)" "$(k -n shop get deploy web -o jsonpath='{.metadata.generation}')"; }
check 1 "Deployment web was not modified" untouched
