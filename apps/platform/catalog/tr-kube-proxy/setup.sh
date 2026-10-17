kubectl -n kube-system patch daemonset kube-proxy --type=merge \
  -p '{"spec":{"template":{"spec":{"nodeSelector":{"kubernetes.io/os":"linux","kube-proxy":"enabled"}}}}}' >/dev/null
gone() { [ -z "$(kubectl -n kube-system get pods -l k8s-app=kube-proxy -o name)" ]; }
wait_for 120 gone

fresh_ns tr-kube-proxy
kubectl -n tr-kube-proxy create deployment web --image=nginx:1.27-alpine --replicas=2 >/dev/null
kubectl -n tr-kube-proxy expose deployment web --port=80 >/dev/null
kubectl -n tr-kube-proxy run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n tr-kube-proxy rollout status deployment/web --timeout=120s >/dev/null
kubectl -n tr-kube-proxy wait --for=condition=Ready pod/client --timeout=120s >/dev/null
