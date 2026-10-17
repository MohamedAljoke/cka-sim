fresh_ns st-fsgroup
fresh_course
kubectl -n st-fsgroup create secret generic db --from-literal=password='hunter2' >/dev/null
cat > "$COURSE/reader.yaml" <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: reader
  namespace: st-fsgroup
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 1000
  containers:
  - name: reader
    image: busybox:1.36
    command: [sh, -c, 'cat /creds/password && echo && sleep 1d']
    volumeMounts:
    - name: creds
      mountPath: /creds
  volumes:
  - name: creds
    secret:
      secretName: db
      defaultMode: 0440
YAML
chmod a+rw "$COURSE/reader.yaml"
kubectl apply -f "$COURSE/reader.yaml" >/dev/null
crashed() { kubectl -n st-fsgroup logs reader 2>/dev/null | grep -q 'Permission denied'; }
wait_for 60 crashed
