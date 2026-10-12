kubectl create priorityclass critical --value=1000000 --description="business-critical workloads"
sed -i 's/^spec:$/spec:\n  priorityClassName: critical/' /opt/course/wl-priority/important.yaml
kubectl apply -f /opt/course/wl-priority/important.yaml
kubectl -n wl-priority wait --for=condition=Ready pod/important --timeout=90s
