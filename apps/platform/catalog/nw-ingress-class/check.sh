default() { kubectl get ingressclass nginx -o jsonpath='{.metadata.annotations.ingressclass\.kubernetes\.io/is-default-class}'; }
served() { eq "legacy /" "$(fetch legacy.nw-ingress-class.test /)"; }
class() { eq nginx "$(kubectl -n nw-ingress-class get ingress legacy -o jsonpath='{.spec.ingressClassName}')"; }
file_kept() { class && ! grep -q ingressClassName "$COURSE/legacy.yaml"; }

check 2 "IngressClass nginx is marked as the default" eq true "$(default)"
check 2 "http://legacy.nw-ingress-class.test reaches legacy" served
check 1 "Ingress legacy got class nginx from the default, and legacy.yaml still has no ingressClassName" file_kept
