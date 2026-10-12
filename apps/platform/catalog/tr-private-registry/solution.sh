kubectl -n tr-private-registry create secret docker-registry registry-creds --docker-server=cka-sim-registry:5000 --docker-username=candidate --docker-password=cka-registry >/dev/null
kubectl -n tr-private-registry patch serviceaccount default -p '{"imagePullSecrets":[{"name":"registry-creds"}]}' >/dev/null
kubectl -n tr-private-registry rollout restart deployment web >/dev/null
kubectl -n tr-private-registry rollout status deployment web --timeout=90s >/dev/null
