# Break the scheduler's static Pod manifest, then create a Pod that can only start once it is fixed.
on_host "sed -i 's|^    - kube-scheduler\$|    - kube-schedulerr|' /etc/kubernetes/manifests/kube-scheduler.yaml"
fresh_ns sched-check
k apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: probe
  namespace: sched-check
spec:
  # Pinned to the control plane (and tolerating its taint) so it never lands on a worker
  # another task may have broken.
  nodeSelector:
    node-role.kubernetes.io/control-plane: ""
  tolerations:
  - operator: Exists
  containers:
  - name: pause
    image: registry.k8s.io/pause:3.10
YAML
remember probe-uid "$(k -n sched-check get pod probe -o jsonpath='{.metadata.uid}')"
