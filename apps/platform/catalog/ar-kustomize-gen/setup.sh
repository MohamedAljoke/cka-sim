fresh_ns ar-kustomize-gen
fresh_course
cat > "$COURSE/app.properties" <<'TXT'
mode=prod
workers=4
TXT
cat > "$COURSE/deployment.yaml" <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: [sh, -c, sleep 1d]
        env:
        - name: PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: password
        volumeMounts:
        - name: config
          mountPath: /etc/app
      volumes:
      - name: config
        configMap:
          name: app-config
YAML
chmod -R a+rwX "$COURSE"
