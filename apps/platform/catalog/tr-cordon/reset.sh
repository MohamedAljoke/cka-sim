[ "$(kubectl get node cka-sim-worker2 -o jsonpath='{.spec.unschedulable}')" = true ] || exit 0
kubectl uncordon cka-sim-worker2 >/dev/null
