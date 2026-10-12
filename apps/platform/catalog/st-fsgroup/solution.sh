# kubectl -n st-fsgroup logs reader -> cat: can't open '/creds/password': Permission denied
f=/opt/course/st-fsgroup/reader.yaml
sed -i 's/^    runAsGroup: 1000$/    runAsGroup: 1000\n    fsGroup: 2000/' "$f"
kubectl replace --force -f "$f"
kubectl -n st-fsgroup wait --for=condition=Ready pod/reader --timeout=60s
