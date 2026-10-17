fresh_ns wl-image-digest
fresh_course
kubectl -n wl-image-digest create secret docker-registry registry-creds --docker-server=cka-sim-registry:5000 --docker-username=candidate --docker-password=cka-registry >/dev/null
kubectl -n wl-image-digest patch serviceaccount default -p '{"imagePullSecrets":[{"name":"registry-creds"}]}' >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: wl-image-digest
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
        image: cka-sim-registry:5000/team/web:1.1
        imagePullPolicy: Always
YAML
kubectl -n wl-image-digest rollout status deployment web --timeout=90s >/dev/null
kubectl -n wl-image-digest get pods -l app=web -o jsonpath='{.items[0].status.containerStatuses[0].imageID}' | sed 's/.*@//' > /etc/cka-sim/wl-image-digest
