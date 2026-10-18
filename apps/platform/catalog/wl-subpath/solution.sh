kubectl -n wl-subpath patch deployment app --type=json -p '[{"op":"replace","path":"/spec/template/spec/containers/0/volumeMounts/0","value":{"name":"config","mountPath":"/etc/app"}}]'
kubectl -n wl-subpath rollout status deployment app --timeout=120s
