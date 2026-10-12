fresh_ns tr-qos
fresh_course
cat > "$COURSE/payments.yaml" <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: payments
  namespace: tr-qos
spec:
  containers:
  - name: api
    image: busybox:1.36
    command: [sleep, 1d]
    resources:
      requests:
        cpu: 100m
        memory: 64Mi
YAML
chmod a+rw "$COURSE/payments.yaml"
kubectl apply -f "$COURSE/payments.yaml" >/dev/null
kubectl -n tr-qos run search --image=busybox:1.36 --overrides='{"spec":{"containers":[{"name":"search","image":"busybox:1.36","command":["sleep","1d"],"resources":{"requests":{"cpu":"50m","memory":"32Mi"},"limits":{"cpu":"50m","memory":"32Mi"}}}]}}' >/dev/null
kubectl -n tr-qos run nightly --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n tr-qos wait --for=condition=Ready pod --all --timeout=120s >/dev/null
