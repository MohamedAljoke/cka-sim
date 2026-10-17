fresh_ns nw-ingress-pathtype
web_app nw-ingress-pathtype docs
kubectl -n nw-ingress-pathtype create ingress docs --class=nginx --rule='docs.nw-ingress-pathtype.test/docs=docs:80' >/dev/null
