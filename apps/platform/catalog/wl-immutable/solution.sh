kubectl -n wl-immutable patch configmap app-config-v1 -p '{"immutable":true}'
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config-v2
  namespace: wl-immutable
immutable: true
data:
  color: blue
YAML
kubectl -n wl-immutable patch deployment app --type=json -p '[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/configMapKeyRef/name","value":"app-config-v2"}]'
kubectl -n wl-immutable rollout status deployment app --timeout=120s
