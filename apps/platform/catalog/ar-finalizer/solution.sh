# kubectl -n ar-finalizer get pvc data -o jsonpath='{.metadata.finalizers}' -> ["kubernetes.io/pvc-protection"]
# kubectl -n ar-finalizer describe pvc data                                 -> Used By: legacy-app
kubectl -n ar-finalizer delete pod legacy-app --grace-period=1
kubectl -n ar-finalizer wait --for=delete pvc/data --timeout=60s
