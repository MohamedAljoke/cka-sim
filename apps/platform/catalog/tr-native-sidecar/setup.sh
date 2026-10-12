fresh_ns tr-native-sidecar
fresh_course
cat > "$COURSE/report.yaml" <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: report
  namespace: tr-native-sidecar
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sh, -c, 'i=0; while true; i=$((i+1)); do echo "report line $i" >> /var/log/report/report.log; sleep 2; done']
    volumeMounts:
    - name: logs
      mountPath: /var/log/report
  volumes:
  - name: logs
    emptyDir: {}
YAML
chmod a+rw "$COURSE/report.yaml"
kubectl apply -f "$COURSE/report.yaml" >/dev/null
kubectl -n tr-native-sidecar wait --for=condition=Ready pod/report --timeout=120s >/dev/null
