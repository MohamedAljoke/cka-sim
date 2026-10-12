fresh_ns ar-helm-broken
fresh_course
chart() { helm -n ar-helm-broken install "$1" web --repo http://cka-sim-charts --version "$2" "${@:3}" >/dev/null 2>&1; }
chart blog 1.0.0
chart wiki 1.1.0
chart cart 1.1.0 --set image.tag=9.9-missing --wait --timeout 15s || true
chart forum 1.0.0
