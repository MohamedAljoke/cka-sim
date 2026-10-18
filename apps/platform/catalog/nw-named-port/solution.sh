kubectl -n nw-named-port create service clusterip api --tcp=80:8080 --dry-run=client -o yaml |
  sed 's/targetPort: 8080/targetPort: http/' | kubectl apply -f -
kubectl -n nw-named-port patch deployment api --type=json -p '[
  {"op":"replace","path":"/spec/template/spec/containers/0/env/0/value","value":"9090"},
  {"op":"replace","path":"/spec/template/spec/containers/0/ports/0/containerPort","value":9090}]'
kubectl -n nw-named-port rollout status deployment api --timeout=120s
