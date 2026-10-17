fresh_ns st-projected
kubectl -n st-projected create configmap app-config --from-literal=settings='mode=fast' >/dev/null
kubectl -n st-projected create secret generic app-secret --from-literal=token='s3cr3t-token' >/dev/null
