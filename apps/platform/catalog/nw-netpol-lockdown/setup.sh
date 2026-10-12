fresh_ns nw-netpol-lockdown
fresh_ns nw-netpol-visitor
kubectl -n nw-netpol-lockdown run app --image=busybox:1.36 --labels=app=app --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-lockdown run web --image=busybox:1.36 --labels=app=web --command -- sh -c 'httpd -p 8080; sleep 1d' >/dev/null
kubectl -n nw-netpol-visitor run visitor --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-lockdown wait --for=condition=Ready pod --all --timeout=120s >/dev/null
kubectl -n nw-netpol-visitor wait --for=condition=Ready pod --all --timeout=120s >/dev/null
