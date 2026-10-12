Finished Jobs and their Pods stay in the cluster until someone deletes them, so a busy namespace
fills up with old Jobs. Two Job fields keep that in check:

| Field | Effect |
|---|---|
| `activeDeadlineSeconds` | after this many seconds from the Job's start, running Pods are killed and the Job fails with reason `DeadlineExceeded`. It beats `backoffLimit`. |
| `ttlSecondsAfterFinished` | the TTL-after-finished controller deletes the Job, and with it its Pods, this many seconds after it completes or fails |

```sh
kubectl -n wl-job-limits create job migrate --image=busybox:1.36 --dry-run=client -o yaml -- sh -c '...' > job.yaml
# add under spec: activeDeadlineSeconds: 60 and ttlSecondsAfterFinished: 300
kubectl apply -f job.yaml
kubectl -n wl-job-limits get job migrate -w
```

Both are **Job**-level fields, under the Job's `spec`, not in the Pod template. The Pod spec has
its own `activeDeadlineSeconds` for a single Pod, which is a common mix-up. For CronJobs, the
history limits usually do the cleanup instead of a TTL.
