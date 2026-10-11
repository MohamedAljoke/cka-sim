manifest=/etc/kubernetes/manifests/kube-scheduler.yaml
grep -q '^    - kube-schedulerr$' "$manifest" || exit 0
sed -i 's|^    - kube-schedulerr$|    - kube-scheduler|' "$manifest"

scheduler_up() { crictl ps -q --name '^kube-scheduler$' | grep -q .; }
wait_for 90 scheduler_up
