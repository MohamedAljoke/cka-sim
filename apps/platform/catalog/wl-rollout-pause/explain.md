Every change to a Deployment's Pod template starts a rollout and creates a new ReplicaSet, which is
one more revision in the history. Two separate edits mean two rollouts, the first one only half
finished when the second starts. **Pausing** lets you stack edits and roll them out once:

```sh
kubectl -n wl-rollout-pause rollout pause deployment web
kubectl -n wl-rollout-pause set image deployment web nginx=nginx:1.27-alpine
kubectl -n wl-rollout-pause set resources deployment web --requests=cpu=50m,memory=32Mi
kubectl -n wl-rollout-pause annotate deployment web kubernetes.io/change-cause="upgrade to 1.27 with requests"
kubectl -n wl-rollout-pause rollout resume deployment web
kubectl -n wl-rollout-pause rollout history deployment web    # CHANGE-CAUSE column
```

While paused, the spec changes but the controller doesn't act on it. `get deploy` shows the old
Pods, and `rollout status` waits. The trap is forgetting `resume`: the Deployment then never
updates.

`--record` is deprecated and gone from most commands. The `kubernetes.io/change-cause`
annotation is what `rollout history` shows, and it's copied to the new ReplicaSet. Set it before
resuming, so it lands on the right revision.
