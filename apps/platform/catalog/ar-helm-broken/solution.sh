helm -n ar-helm-broken ls --failed -q | xargs -r helm -n ar-helm-broken uninstall >/dev/null
helm show values web --repo http://cka-sim-charts --version 1.1.0 > /opt/course/ar-helm-broken/values.yaml
