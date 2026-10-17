fresh_ns nw-netpol-ipblock
fresh_course
on() { printf '{"spec":{"nodeName":"%s"}}' "$1"; }
kubectl -n nw-netpol-ipblock run web --image=busybox:1.36 --labels=app=web --overrides="$(on cka-sim-worker)" --command -- sh -c 'httpd -p 80; sleep 1d' >/dev/null
kubectl -n nw-netpol-ipblock run near --image=busybox:1.36 --overrides="$(on cka-sim-worker2)" --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-ipblock run banned --image=busybox:1.36 --overrides="$(on cka-sim-worker2)" --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-ipblock run far --image=busybox:1.36 --overrides="$(on cka-sim-worker)" --command -- sleep 1d >/dev/null
kubectl -n nw-netpol-ipblock wait --for=condition=Ready pod --all --timeout=120s >/dev/null
kubectl get node cka-sim-worker2 -o jsonpath='{.spec.podCIDR}{"\n"}' > "$COURSE/cidr"
echo "$(pod_ip nw-netpol-ipblock banned)" > "$COURSE/except"
