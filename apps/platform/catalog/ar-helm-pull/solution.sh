helm pull web --repo http://cka-sim-charts --version 1.1.0 --untar --untardir /opt/course/ar-helm-pull
sed -i 's/^replicaCount:.*/replicaCount: 4/' /opt/course/ar-helm-pull/web/values.yaml
helm -n ar-helm-local install local /opt/course/ar-helm-pull/web --create-namespace --wait >/dev/null
