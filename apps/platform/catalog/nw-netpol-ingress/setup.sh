fresh_ns nw-netpol-ingress
fresh_ns nw-netpol-monitor
fresh_ns nw-netpol-outside
kubectl -n nw-netpol-ingress run api --image=busybox:1.36 --labels=app=api --command -- sh -c 'httpd -p 8080; httpd -p 9090; sleep 1d' >/dev/null
kubectl -n nw-netpol-ingress run frontend --image=busybox:1.36 --labels=role=frontend --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-ingress run worker --image=busybox:1.36 --labels=role=worker --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-monitor run probe --image=busybox:1.36 --labels=role=probe --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-outside run stranger --image=busybox:1.36 --labels=role=frontend --command -- sleep 1d >/dev/null
for ns in nw-netpol-ingress nw-netpol-monitor nw-netpol-outside; do
  kubectl -n $ns wait --for=condition=Ready pod --all --timeout=120s >/dev/null
done
