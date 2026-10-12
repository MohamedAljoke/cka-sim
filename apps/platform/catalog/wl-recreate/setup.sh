fresh_ns wl-recreate
kubectl -n wl-recreate create deployment db-migrator --image=nginx:1.27-alpine --replicas=2 >/dev/null
kubectl -n wl-recreate patch deployment db-migrator --type=json -p '[{"op":"add","path":"/spec/template/spec/containers/0/command","value":["sh","-c","sleep 1d"]}]' >/dev/null
kubectl -n wl-recreate rollout status deployment/db-migrator --timeout=120s >/dev/null
