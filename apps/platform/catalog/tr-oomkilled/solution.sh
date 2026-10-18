# kubectl -n tr-oomkilled describe pod -l app=cache -> Last State: Terminated, Reason: OOMKilled, Exit Code: 137
kubectl -n tr-oomkilled get pods -l app=cache -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}' > /opt/course/tr-oomkilled/reason.txt
kubectl -n tr-oomkilled set resources deployment cache --limits=memory=128Mi
kubectl -n tr-oomkilled rollout status deployment cache --timeout=90s
