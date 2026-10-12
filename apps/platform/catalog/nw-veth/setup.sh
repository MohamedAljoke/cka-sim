fresh_ns nw-veth
fresh_course
kubectl -n nw-veth run probe --image=busybox:1.36 --overrides='{"spec":{"nodeName":"cka-sim-worker2"}}' --command -- sleep 1d >/dev/null
kubectl -n nw-veth wait --for=condition=Ready pod/probe --timeout=120s >/dev/null
