fresh_ns nw-gateway-migrate
dir=$(mktemp -d)
openssl req -x509 -newkey rsa:2048 -nodes -days 30 -subj /CN=shop.nw-gateway-migrate.test \
  -addext subjectAltName=DNS:shop.nw-gateway-migrate.test -keyout "$dir/tls.key" -out "$dir/tls.crt" 2>/dev/null
kubectl -n nw-gateway-migrate create secret tls shop-tls --cert="$dir/tls.crt" --key="$dir/tls.key" >/dev/null
rm -rf "$dir"
web_app nw-gateway-migrate store
web_app nw-gateway-migrate api
kubectl -n nw-gateway-migrate create ingress shop --class=nginx \
  --rule='shop.nw-gateway-migrate.test/api*=api:80,tls=shop-tls' \
  --rule='shop.nw-gateway-migrate.test/*=store:80,tls=shop-tls' >/dev/null
