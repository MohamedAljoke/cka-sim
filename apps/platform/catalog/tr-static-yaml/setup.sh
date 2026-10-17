for f in /etc/kubernetes/manifests/*; do
  grep -q 'name: node-exporter' "$f" 2>/dev/null && rm -f "$f"
done
fresh_ns tr-static-yaml
cat > /etc/kubernetes/manifests/node-exporter.yaml <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: node-exporter
  namespace: tr-static-yaml
spec:
  containers:
  - name: exporter
    image: busybox:1.36
     command: [sh, -c, 'while true; do cat /proc/loadavg; sleep 10; done']
YAML
logged() { journalctl -u kubelet --since "-5min" --no-pager | grep -q 'node-exporter.yaml'; }
wait_for 60 logged
