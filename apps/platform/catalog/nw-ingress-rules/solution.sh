kubectl -n nw-ingress-rules create ingress shop --class=nginx \
  --rule='shop.nw-ingress-rules.test/api*=api:80' \
  --rule='shop.nw-ingress-rules.test/*=store:80'
sleep 5
curl -s -H 'Host: shop.nw-ingress-rules.test' http://10.96.80.80/api/orders > /opt/course/nw-ingress-rules/curl.txt
