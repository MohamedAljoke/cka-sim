manifest=/etc/kubernetes/manifests/kube-scheduler.yaml
sed -i 's|^    - kube-schedulerr$|    - kube-scheduler|' "$manifest"
sed -i 's|^    - kube-scheduler$|    - kube-schedulerr|' "$manifest"

fresh_ns sched-check
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: probe
  namespace: sched-check
spec:
  # Pinned to the control plane so it never lands on a worker another task may have broken.
  nodeSelector:
    node-role.kubernetes.io/control-plane: ""
  tolerations:
  - operator: Exists
  containers:
  - name: pause
    image: registry.k8s.io/pause:3.10
YAML

scheduler_down() { ! crictl ps -q --name '^kube-scheduler$' | grep -q .; }
wait_for 90 scheduler_down
