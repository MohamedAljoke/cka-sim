# kubectl -n nw-readiness get pods                -> 0/1 Running: not Ready
# kubectl -n nw-readiness describe pod -l app=web -> Readiness probe failed: HTTP probe failed with statuscode: 404
kubectl -n nw-readiness patch deployment web --type=json \
  -p '[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/"}]'
kubectl -n nw-readiness rollout status deployment web --timeout=120s
