A rolling update first creates **extra** Pods (`maxSurge`, 25% rounded up, so 1 here) before
removing old ones. Two running Pods at 400m plus one surge Pod at 400m is 1200m of requests, over
the quota's `requests.cpu: 1`. The quota admission plugin refuses the new Pod, and the rollout
waits until `progressDeadlineSeconds` and reports `ProgressDeadlineExceeded`.

The Deployment and its Pods show nothing useful. The error is on the **new ReplicaSet**, because
it is the one that tried to create the Pod:

```sh
kubectl -n wl-quota-rollout rollout status deploy api
kubectl -n wl-quota-rollout get rs                       # new RS: DESIRED 1, CURRENT 0
kubectl -n wl-quota-rollout describe rs <new-rs>         # FailedCreate ... exceeded quota: team-cpu
kubectl -n wl-quota-rollout describe quota team-cpu      # Used 800m / Hard 1
```

With the quota, replicas and requests all fixed, the only freedom left is **how** to roll:
`maxSurge: 0, maxUnavailable: 1` removes an old Pod first, then starts its replacement, so the
total never exceeds 800m. `Recreate` would work too, but with downtime.

The rollout continues by itself once the strategy changes. `ProgressDeadlineExceeded` only
reports the problem; it doesn't stop the controller.
