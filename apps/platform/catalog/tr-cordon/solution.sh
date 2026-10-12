# kubectl -n tr-cordon describe pod -l app=agent -> 0/3 nodes available: 1 node(s) were unschedulable, ...
# kubectl get nodes                               -> cka-sim-worker2   Ready,SchedulingDisabled
kubectl uncordon cka-sim-worker2
echo cka-sim-worker2 > /opt/course/tr-cordon/node
