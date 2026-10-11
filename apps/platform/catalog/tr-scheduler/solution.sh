# kubectl -n kube-system get pods  -> kube-scheduler-... in CrashLoopBackOff
# crictl ps -a; crictl logs <id>   -> exec: "kube-schedulerr": executable file not found
# The scheduler is a static Pod: its source is a file on the control plane node.
sed -i 's|^    - kube-schedulerr$|    - kube-scheduler|' /etc/kubernetes/manifests/kube-scheduler.yaml
