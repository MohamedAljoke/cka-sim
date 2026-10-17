fresh_ns nw-gateway-https
dir=$(mktemp -d)
openssl req -x509 -newkey rsa:2048 -nodes -days 30 -subj /CN=secure.nw-gateway-https.test \
  -addext subjectAltName=DNS:secure.nw-gateway-https.test -keyout "$dir/tls.key" -out "$dir/tls.crt" 2>/dev/null
kubectl -n nw-gateway-https create secret tls secure-tls --cert="$dir/tls.crt" --key="$dir/tls.key" >/dev/null
rm -rf "$dir"
web_app nw-gateway-https app
gateway nw-gateway-https secure-gw
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app
  namespace: nw-gateway-https
spec:
  parentRefs:
  - name: secure-gw
  hostnames:
  - secure.nw-gateway-https.test
  rules:
  - backendRefs:
    - name: app
      port: 80
YAML
