fresh_ns tr-container-logs
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: checkout
  namespace: tr-container-logs
spec:
  containers:
  - name: proxy
    image: busybox:1.36
    command:
    - sh
    - -c
    - |
      echo "INFO GET /checkout 200"
      echo "ERROR upstream connect error: 503"
      echo "INFO GET /health 200"
      echo "ERROR upstream reset before headers"
      sleep 1d
  - name: app
    image: busybox:1.36
    command:
    - sh
    - -c
    - |
      echo "INFO order=1001 received"
      echo "INFO order=1001 paid"
      echo "ERROR order=1002 payment declined: card expired"
      echo "INFO order=1003 received"
      echo "ERROR order=1004 inventory lookup timed out"
      echo "INFO order=1003 paid"
      echo "ERROR order=1005 payment declined: insufficient funds"
      sleep 1d
YAML
kubectl -n tr-container-logs wait --for=condition=Ready pod/checkout --timeout=120s >/dev/null
