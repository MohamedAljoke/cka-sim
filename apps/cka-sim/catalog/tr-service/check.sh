endpoints() { [ -n "$(kubectl -n shop get endpointslices -l kubernetes.io/service-name=web -o jsonpath='{.items[*].endpoints[*].addresses[*]}')" ]; }
serves() { [[ "$(kubectl -n shop exec deploy/web -- wget -qO- -T 3 http://web.shop 2>/dev/null)" == *"Welcome to nginx"* ]]; }
# setup.sh creates the Deployment fresh, so any edit to its spec (even a scale) raises this above 1.
generation() { kubectl -n shop get deployment web -o jsonpath='{.metadata.generation}'; }
untouched() { serves && eq 1 "$(generation)"; }

wait_for 20 serves
check 2 "Service web has endpoints" endpoints
check 4 "http://web.shop returns the nginx page" serves
check 1 "Deployment web was not modified" untouched
