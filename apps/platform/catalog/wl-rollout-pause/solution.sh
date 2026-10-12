kubectl -n wl-rollout-pause rollout pause deployment web
kubectl -n wl-rollout-pause set image deployment web nginx=nginx:1.27-alpine
kubectl -n wl-rollout-pause set resources deployment web --requests=cpu=50m,memory=32Mi
kubectl -n wl-rollout-pause annotate deployment web kubernetes.io/change-cause="upgrade to 1.27 with requests"
kubectl -n wl-rollout-pause rollout resume deployment web
kubectl -n wl-rollout-pause rollout status deployment web --timeout=120s
kubectl -n wl-rollout-pause rollout history deployment web
