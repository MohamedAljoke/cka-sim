fresh_ns tr-top-containers
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: tr-top-containers
spec:
  containers:
  - name: web
    image: nginx:1.27-alpine
  - name: cache
    image: busybox:1.36
    command: [sh, -c, 'x=$(head -c 50000000 /dev/zero | tr "\0" a); sleep 1d']
    resources:
      limits: {memory: 256Mi}
  - name: logger
    image: busybox:1.36
    command: [sleep, 1d]
YAML
measured() { kubectl -n tr-top-containers top pod app --containers --no-headers 2>/dev/null | awk '$2 == "cache" && $4 + 0 >= 40 { full = 1 } END { exit !(full && NR == 3) }'; }
wait_for 150 measured
