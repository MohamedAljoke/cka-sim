fresh_ns wl-hpa-fight
fresh_course
cat > "$COURSE/web.yaml" <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: wl-hpa-fight
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: web
        image: nginx:1.27-alpine
        resources:
          requests:
            cpu: 50m
YAML
chmod 0666 "$COURSE/web.yaml"
kubectl apply -f "$COURSE/web.yaml" >/dev/null
kubectl -n wl-hpa-fight autoscale deployment web --min=3 --max=6 --cpu=80% >/dev/null
replicas() { eq "$1" "$(kubectl -n wl-hpa-fight get deployment web -o jsonpath='{.spec.replicas}')"; }
wait_for 60 replicas 3
kubectl apply -f "$COURSE/web.yaml" >/dev/null
