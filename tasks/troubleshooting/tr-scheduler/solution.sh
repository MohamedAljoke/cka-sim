# k -n kube-system get pods            -> kube-scheduler-... in CrashLoopBackOff / Error
# k -n kube-system logs kube-scheduler-cka3962-control-plane  (or: crictl ps -a; crictl logs <id>)
#   -> exec: "kube-schedulerr": executable file not found
# The scheduler is a static Pod: its source is a file on the control plane node.
on_host "sed -i 's|^    - kube-schedulerr\$|    - kube-scheduler|' /etc/kubernetes/manifests/kube-scheduler.yaml"
