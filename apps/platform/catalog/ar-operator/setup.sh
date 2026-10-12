kubectl delete -f /etc/cka-sim/addons/web-operator.yaml --ignore-not-found --wait=true --timeout=120s >/dev/null 2>&1
fresh_ns ar-operator
fresh_course
cp /etc/cka-sim/addons/web-operator.yaml "$COURSE/web-operator.yaml"
