# kubectl drain cka-sim-worker2 -> refuses: DaemonSet-managed Pods, Pods with local storage, Pods without a controller
kubectl drain cka-sim-worker2 --ignore-daemonsets --delete-emptydir-data --force --timeout=120s
