kubectl -n nw-gateway-https patch gateway secure-gw --type=json -p '[{"op":"add","path":"/spec/listeners/-","value":{"name":"https","protocol":"HTTPS","port":443,"hostname":"secure.nw-gateway-https.test","tls":{"mode":"Terminate","certificateRefs":[{"name":"secure-tls"}]}}}]'
kubectl -n nw-gateway-https patch httproute app --type=json -p '[{"op":"add","path":"/spec/parentRefs/0/sectionName","value":"https"}]'
kubectl apply -f - <<'YAML'
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: redirect
  namespace: nw-gateway-https
spec:
  parentRefs:
  - name: secure-gw
    sectionName: http
  hostnames:
  - secure.nw-gateway-https.test
  rules:
  - filters:
    - type: RequestRedirect
      requestRedirect:
        scheme: https
        statusCode: 301
YAML
sleep 5
