kubectl apply -f /opt/course/ar-operator/web-operator.yaml
kubectl wait --for=condition=Established crd/webapps.cka-sim.io --timeout=60s
kubectl -n web-operator rollout status deployment web-operator --timeout=120s
kubectl apply -f - <<'YAML'
apiVersion: cka-sim.io/v1
kind: WebApp
metadata:
  name: shop
  namespace: ar-operator
spec:
  image: nginx:1.27-alpine
  replicas: 2
  port: 80
YAML
kubectl -n ar-operator wait --for=jsonpath='{.status.readyReplicas}'=2 webapp/shop --timeout=120s
