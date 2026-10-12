helm repo add cka http://cka-sim-charts >/dev/null
helm repo update cka >/dev/null
helm -n ar-helm-install install shop cka/web --version 1.0.0 --set replicaCount=3 --wait >/dev/null
helm -n ar-helm-install upgrade shop cka/web --version 1.1.0 --reuse-values --wait >/dev/null
helm -n ar-helm-install rollback shop 1 --wait >/dev/null
