releases() { helm -n ar-helm-broken ls -a -q 2>/dev/null | sort | xargs; }
cart_gone() { ! kubectl -n ar-helm-broken get deployment cart-web >/dev/null 2>&1; }
same_values() { diff -q <(helm show values web --repo http://cka-sim-charts --version 1.1.0) "$COURSE/values.yaml"; }

check 2 "release cart is uninstalled, blog, forum and wiki remain" eq "blog forum wiki" "$(releases)"
check 1 "cart's Deployment is gone" cart_gone
check 2 "values.yaml holds the default values of web 1.1.0" same_values
