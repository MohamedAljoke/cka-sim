kubectl -n wl-recreate patch deployment db-migrator --type=merge -p '{"spec":{"strategy":{"type":"Recreate","rollingUpdate":null}}}'
kubectl -n wl-recreate set image deployment db-migrator nginx=busybox:1.36
kubectl -n wl-recreate rollout status deployment db-migrator --timeout=120s
