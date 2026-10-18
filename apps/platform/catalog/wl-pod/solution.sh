kubectl -n wl-pod run web --image=nginx:1.27
kubectl -n wl-pod wait --for=condition=Ready pod/web --timeout=120s
