fresh_ns tr-events
fresh_course
pod() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: v1
kind: Pod
metadata:
  name: $1
  namespace: tr-events
spec:
  containers:
  - name: main
    image: busybox:1.36
    command: [sh, -c, sleep 1d]
    volumeMounts:
    - name: config
      mountPath: /config
  volumes:
  - name: config
    $2
YAML
}
pod report 'configMap: {name: report-config}'
pod ledger 'secret: {secretName: ledger-credentials}'
failed_mount() { [ -n "$(kubectl -n tr-events get events --field-selector involvedObject.name="$1",reason=FailedMount -o name)" ]; }
wait_for 90 failed_mount report
wait_for 30 failed_mount ledger
