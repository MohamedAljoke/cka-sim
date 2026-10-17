fresh_ns tr-log-agent
fresh_ns tr-log-agent-apps
kubectl -n tr-log-agent-apps run app --image=busybox:1.36 --command -- sh -c 'i=0; while true; i=$((i+1)); do echo "app heartbeat $i"; sleep 2; done' >/dev/null
kubectl -n tr-log-agent-apps wait --for=condition=Ready pod/app --timeout=120s >/dev/null
