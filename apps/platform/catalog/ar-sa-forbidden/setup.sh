fresh_ns ar-sa
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: reporter
  namespace: ar-sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
  namespace: ar-sa
rules:
- apiGroups: [""]
  resources: [pods]
  verbs: [get, watch]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: reporter-pod-reader
  namespace: ar-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pod-reader
subjects:
- kind: ServiceAccount
  name: reporter
  namespace: ar-sa
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reporter
  namespace: ar-sa
spec:
  replicas: 1
  selector:
    matchLabels:
      app: reporter
  template:
    metadata:
      labels:
        app: reporter
    spec:
      serviceAccountName: reporter
      containers:
      - name: reporter
        image: busybox:1.36
        command: [sh, -c, sleep 1d]
YAML
