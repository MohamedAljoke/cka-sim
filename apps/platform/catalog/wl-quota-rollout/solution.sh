# kubectl -n wl-quota-rollout describe deploy api  -> Progressing False ProgressDeadlineExceeded
# kubectl -n wl-quota-rollout get rs               -> the new ReplicaSet has DESIRED 1, CURRENT 0
rs=$(kubectl -n wl-quota-rollout get rs -l app=api --sort-by=.metadata.creationTimestamp -o name | tail -1)
kubectl -n wl-quota-rollout get events --field-selector involvedObject.name="${rs#*/}",reason=FailedCreate -o jsonpath='{.items[-1].message}' > /opt/course/wl-quota-rollout/reason.txt
kubectl -n wl-quota-rollout patch deployment api -p '{"spec":{"strategy":{"rollingUpdate":{"maxSurge":0,"maxUnavailable":1}}}}'
kubectl -n wl-quota-rollout rollout status deployment api --timeout=120s
