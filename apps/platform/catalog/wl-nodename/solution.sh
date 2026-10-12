kubectl -n wl-nodename run pinned --image=nginx:1.27-alpine --overrides='{"spec":{"nodeName":"cka-sim-worker2"}}'
# grep staticPodPath /var/lib/kubelet/config.yaml -> /etc/kubernetes/manifests
kubectl -n wl-nodename run static-web --image=nginx:1.27-alpine --dry-run=client -o yaml > /etc/kubernetes/manifests/static-web.yaml
