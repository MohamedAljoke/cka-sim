When a container uses more memory than its **limit**, the kernel's OOM killer ends it with
`SIGKILL`. Kubernetes records `Reason: OOMKilled` and `Exit Code: 137` (128 + 9), then restarts it,
and soon you see `CrashLoopBackOff`. The app never gets to log anything, which is why the logs look
normal. The reason is only in the container's status:

```sh
kubectl -n tr-oomkilled get pods                         # RESTARTS climbing
kubectl -n tr-oomkilled describe pod -l app=cache        # Last State: Terminated  Reason: OOMKilled
kubectl -n tr-oomkilled get pod <pod> -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}'
kubectl -n tr-oomkilled set resources deployment cache --limits=memory=128Mi
```

Tell it apart from an app crash: exit code 1 or 2 with an error in `logs --previous` is the app.
137 with `OOMKilled` is the limit. Fix it by raising the limit above the real peak (or fixing the
app's memory use), not by removing the limit.

**Requests** are what the scheduler reserves on a node. **Limits** are what the kernel enforces.
Memory over the limit kills the container. CPU over the limit is only throttled.
