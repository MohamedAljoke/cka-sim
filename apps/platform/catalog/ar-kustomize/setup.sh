fresh_ns ar-kustomize
fresh_course
mkdir -p "$COURSE/base" "$COURSE/overlays/prod"
cat > "$COURSE/base/deployment.yaml" <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shop
  labels:
    app: shop
spec:
  replicas: 1
  selector:
    matchLabels:
      app: shop
  template:
    metadata:
      labels:
        app: shop
    spec:
      containers:
      - name: shop
        image: nginx:1.26-alpine
YAML
cat > "$COURSE/base/kustomization.yaml" <<'YAML'
resources:
- deployment.yaml
YAML
chmod -R a+rwX "$COURSE"
