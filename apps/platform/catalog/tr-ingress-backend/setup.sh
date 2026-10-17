fresh_ns tr-ingress-backend
web_app tr-ingress-backend storefront
web_app tr-ingress-backend api
kubectl -n tr-ingress-backend create ingress store --class=nginx \
  --rule='store.tr-ingress-backend.test/*=storefront:8080' \
  --rule='store.tr-ingress-backend.test/api*=store-api:80' >/dev/null
