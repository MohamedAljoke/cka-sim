fresh_ns ar-helm-reuse
chart() { helm -n ar-helm-reuse "$1" portal web --repo http://cka-sim-charts --version 1.1.0 "${@:2}" >/dev/null; }
chart install --set replicaCount=3 --set service.type=NodePort
chart upgrade --set image.pullPolicy=Always
