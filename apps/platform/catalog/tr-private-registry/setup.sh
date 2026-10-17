fresh_ns tr-private-registry
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: tr-private-registry
spec:
  replicas: 2
  selector:
    matchLabels: {app: web}
  template:
    metadata:
      labels: {app: web}
    spec:
      containers:
      - name: web
        image: cka-sim-registry:5000/team/web:1.0
        imagePullPolicy: Always
YAML
