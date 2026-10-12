kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Secret
metadata:
  name: deployer-token
  namespace: ar-legacy-token
  annotations:
    kubernetes.io/service-account.name: deployer
type: kubernetes.io/service-account-token
YAML
until [ -n "$(kubectl -n ar-legacy-token get secret deployer-token -o jsonpath='{.data.token}')" ]; do sleep 1; done
kubectl -n ar-legacy-token get secret deployer-token -o jsonpath='{.data.token}' | base64 -d > /opt/course/ar-legacy-token/token
