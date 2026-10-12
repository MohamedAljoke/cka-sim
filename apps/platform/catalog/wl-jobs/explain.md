A **Job** runs Pods until a number of them succeed:

| Field | Meaning |
|---|---|
| `completions` | how many successful Pods are needed in total |
| `parallelism` | how many run at the same time |
| `backoffLimit` | how many retries before the Job is marked Failed (default 6) |
| `activeDeadlineSeconds` | a time limit for the whole Job |

A **CronJob** creates a Job on a cron schedule (`*/15 * * * *` is every 15 minutes, in the
controller-manager's time zone unless `timeZone` is set):

| Field | Meaning |
|---|---|
| `successfulJobsHistoryLimit` / `failedJobsHistoryLimit` | how many finished Jobs to keep (defaults 3 and 1) |
| `concurrencyPolicy` | `Allow` (default), `Forbid` (skip a run while one is active), `Replace` |
| `suspend` | `true` pauses the schedule |

```sh
kubectl -n wl-jobs create job batch --image=busybox:1.36 --dry-run=client -o yaml -- echo processed > job.yaml
kubectl -n wl-jobs create cronjob report --image=busybox:1.36 --schedule='*/15 * * * *' --dry-run=client -o yaml -- date > cj.yaml
# add the fields above, then kubectl apply -f
kubectl -n wl-jobs get jobs,pods
kubectl -n wl-jobs create job manual-run --from=cronjob/report      # test a CronJob without waiting
```

The imperative commands have no flags for these fields, so `--dry-run=client -o yaml`, edit and
apply is the exam-speed path. A Job's Pods need `restartPolicy: Never` or `OnFailure`; the
generator sets `Never`.
