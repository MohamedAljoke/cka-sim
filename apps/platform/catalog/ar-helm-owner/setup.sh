fresh_ns ar-helm-owner
fresh_ns ar-helm-owner-audit
fresh_course
chart() { helm -n "$1" install "$2" web --repo http://cka-sim-charts --version 1.0.0 --set fullnameOverride="$3" >/dev/null; }
chart ar-helm-owner storefront cart
chart ar-helm-owner payments checkout
chart ar-helm-owner inventory catalog
chart ar-helm-owner-audit audit-trail logs
