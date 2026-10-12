fresh_ns nw-ingress-default
web_app nw-ingress-default site
web_app nw-ingress-default fallback
kubectl -n nw-ingress-default create ingress site --class=nginx --rule='www.nw-ingress-default.test/*=site:80' >/dev/null
