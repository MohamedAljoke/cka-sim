kubectl apply -f - <<'YAML'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: batch-runner
  namespace: ar-no-automount
automountServiceAccountToken: false
---
apiVersion: v1
kind: Pod
metadata:
  name: runner
  namespace: ar-no-automount
spec:
  serviceAccountName: batch-runner
  automountServiceAccountToken: false
  containers:
  - name: runner
    image: busybox:1.36
    command: [sleep, 1d]
YAML
kubectl -n ar-no-automount wait --for=condition=Ready pod/runner --timeout=60s
