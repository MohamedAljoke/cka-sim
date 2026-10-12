fresh_ns nw-ingress-rewrite
web_app nw-ingress-rewrite web
kubectl -n nw-ingress-rewrite create ingress app --class=nginx --rule='app.nw-ingress-rewrite.test/app=web:80' >/dev/null
