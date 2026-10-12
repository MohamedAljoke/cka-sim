fresh_ns st-hostpath-type
fresh_course
rm -rf /var/lib/agent-data
cat > "$COURSE/agent.yaml" <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: agent
  namespace: st-hostpath-type
spec:
  nodeName: cka-sim-control-plane
  containers:
  - name: agent
    image: busybox:1.36
    command: [sh, -c, 'date >> /state/starts; sleep 1d']
    volumeMounts:
    - name: state
      mountPath: /state
  volumes:
  - name: state
    hostPath:
      path: /var/lib/agent-data
      type: Directory
YAML
chmod a+rw "$COURSE/agent.yaml"
kubectl apply -f "$COURSE/agent.yaml" >/dev/null
failed() { kubectl -n st-hostpath-type get events --field-selector involvedObject.name=agent,reason=FailedMount -o name | grep -q .; }
wait_for 60 failed
