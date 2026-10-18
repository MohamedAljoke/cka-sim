kubectl -n tr-debug debug app --image=busybox:1.36 --container=debugger --target=app -- sh -c 'sleep 1d'
until [ -n "$(kubectl -n tr-debug get pod app -o jsonpath='{.status.ephemeralContainerStatuses[0].state.running}')" ]; do sleep 1; done
kubectl -n tr-debug exec app -c debugger -- ps > /opt/course/tr-debug/ps.txt
