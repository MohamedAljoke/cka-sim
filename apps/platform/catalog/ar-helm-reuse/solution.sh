helm -n ar-helm-reuse upgrade portal web --repo http://cka-sim-charts --version 1.1.0 --reuse-values --set replicaCount=3 --set service.type=NodePort --wait >/dev/null
