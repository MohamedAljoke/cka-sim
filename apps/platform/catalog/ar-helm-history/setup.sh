fresh_ns ar-helm-history
fresh_course
chart() { helm -n ar-helm-history "$1" api web --repo http://cka-sim-charts --version "$2" "${@:3}" >/dev/null; }
chart install 1.0.0 --set replicaCount=2
chart upgrade 1.1.0 --reuse-values --set service.type=NodePort
chart upgrade 1.1.0 --set replicaCount=1
