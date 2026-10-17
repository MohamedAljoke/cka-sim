fresh_ns nw-netpol-egress
kubectl -n nw-netpol-egress run backend --image=busybox:1.36 --labels=app=backend --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-egress run frontend --image=busybox:1.36 --labels=app=frontend --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-egress run db --image=busybox:1.36 --labels=app=db --command -- sh -c 'httpd -p 5432; httpd -p 3306; sleep 1d' >/dev/null
kubectl -n nw-netpol-egress run cache --image=busybox:1.36 --labels=app=cache --command -- sh -c 'httpd -p 6379; sleep 1d' >/dev/null
kubectl -n nw-netpol-egress wait --for=condition=Ready pod --all --timeout=120s >/dev/null
