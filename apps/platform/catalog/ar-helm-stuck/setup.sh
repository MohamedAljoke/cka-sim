fresh_ns ar-helm-stuck
chart() { helm -n ar-helm-stuck "$1" billing web --repo http://cka-sim-charts "${@:2}"; }
chart install --version 1.0.0 --wait --timeout 90s >/dev/null
chart upgrade --version 1.1.0 --set image.tag=9.9-missing --wait --timeout 10m </dev/null >/dev/null 2>&1 &
upgrade=$!
pending() { eq pending-upgrade "$(kubectl -n ar-helm-stuck get secret sh.helm.release.v1.billing.v2 -o jsonpath='{.metadata.labels.status}')"; }
wait_for 60 pending
sleep 3
kill -9 "$upgrade"
wait "$upgrade" 2>/dev/null
true
