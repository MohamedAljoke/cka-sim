kubectl apply -f - <<'YAML'
apiVersion: batch/v1
kind: Job
metadata:
  name: importer
  namespace: wl-pod-failure-policy
spec:
  backoffLimit: 6
  podFailurePolicy:
    rules:
    - action: FailJob
      onExitCodes:
        containerName: importer
        operator: In
        values: [42]
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: importer
        image: busybox:1.36
        command: [sh, -c, 'echo invalid input; exit 42']
YAML
kubectl -n wl-pod-failure-policy wait --for=condition=Failed job/importer --timeout=60s
