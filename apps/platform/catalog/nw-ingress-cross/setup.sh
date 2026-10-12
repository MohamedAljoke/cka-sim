fresh_ns nw-ingress-cross
fresh_ns nw-ingress-cross-billing
web_app nw-ingress-cross portal
web_app nw-ingress-cross-billing billing
kubectl -n nw-ingress-cross create ingress portal --class=nginx \
  --rule='portal.nw-ingress-cross.test/*=portal:80' \
  --rule='portal.nw-ingress-cross.test/billing*=billing:80' >/dev/null
