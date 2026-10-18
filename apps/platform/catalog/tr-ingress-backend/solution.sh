# kubectl -n tr-ingress-backend describe ingress store   -> store-api:80 (<error: services "store-api" not found>)
# kubectl -n tr-ingress-backend get svc                   -> storefront listens on 80, not 8080
kubectl -n tr-ingress-backend patch ingress store --type=json -p '[
  {"op":"replace","path":"/spec/rules/0/http/paths/0/backend/service/port/number","value":80},
  {"op":"replace","path":"/spec/rules/0/http/paths/1/backend/service/name","value":"api"}]'
sleep 5
