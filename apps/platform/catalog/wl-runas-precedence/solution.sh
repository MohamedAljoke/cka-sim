kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: two
  namespace: wl-runas-precedence
spec:
  securityContext:
    runAsUser: 1000
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
  - name: sidecar
    image: busybox:1.36
    command: [sleep, 1d]
    securityContext:
      runAsUser: 2000
YAML
kubectl -n wl-runas-precedence wait --for=condition=Ready pod/two --timeout=60s
printf 'app: %s\nsidecar: %s\n' "$(kubectl -n wl-runas-precedence exec two -c app -- id -u)" "$(kubectl -n wl-runas-precedence exec two -c sidecar -- id -u)" > /opt/course/wl-runas-precedence/uids.txt
