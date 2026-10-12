fresh_ns tr-top-pod
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: atlas
  namespace: tr-top-pod
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sh, -c, 'x=$(head -c 60000000 /dev/zero | tr "\0" a); sleep 1d']
    resources:
      limits: {memory: 256Mi}
---
apiVersion: v1
kind: Pod
metadata:
  name: boreas
  namespace: tr-top-pod
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sh, -c, 'while :; do :; done']
    resources:
      limits: {cpu: 150m}
---
apiVersion: v1
kind: Pod
metadata:
  name: cetus
  namespace: tr-top-pod
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
YAML
usage() { kubectl -n tr-top-pod top pod "$1" --no-headers 2>/dev/null | awk -v col="$2" -v min="$3" '{ ok = $col + 0 >= min } END { exit !ok }'; }
measured() { usage atlas 3 40 && usage boreas 2 50 && usage cetus 2 0; }
wait_for 150 measured
