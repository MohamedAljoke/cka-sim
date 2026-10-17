conf=/etc/cni/net.d/10-kindnet.conflist
sed -i 's/"type": "ptpp"/"type": "ptp"/' "$conf"
sed -i 's/"type": "ptp"/"type": "ptpp"/' "$conf"

fresh_ns tr-cni
kubectl -n tr-cni create deployment web --image=nginx:1.27-alpine --dry-run=client -o yaml |
  sed 's|^    spec:$|    spec:\n      nodeSelector:\n        kubernetes.io/hostname: cka-sim-worker2|' |
  kubectl apply -f - >/dev/null
sandbox_failed() { kubectl -n tr-cni get events --field-selector reason=FailedCreatePodSandBox -o name | grep -q .; }
wait_for 90 sandbox_failed
