kubectl -n nw-ingress-cross-billing create ingress billing --class=nginx \
  --rule='portal.nw-ingress-cross.test/billing*=billing:80'
kubectl -n nw-ingress-cross patch ingress portal --type=json -p '[{"op":"remove","path":"/spec/rules/0/http/paths/1"}]'
sleep 5
