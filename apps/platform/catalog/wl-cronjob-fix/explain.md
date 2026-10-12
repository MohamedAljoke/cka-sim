A CronJob that never fires has a short list of suspects. `kubectl get cronjob` shows most of them
in one line:

```sh
kubectl -n wl-cronjob-fix get cronjob cleanup          # SCHEDULE, SUSPEND, ACTIVE, LAST SCHEDULE
kubectl -n wl-cronjob-fix describe cronjob cleanup     # events: missed schedules, failures
```

1. **`suspend: true`** pauses it. No Jobs are created until it's set back to `false`.
2. **A schedule that never matches.** `0 0 31 2 *` is valid cron but means February 31st. The five
   fields are minute, hour, day of month, month, day of week. `*/10 * * * *` means every 10
   minutes.
3. **`concurrencyPolicy: Forbid`** with a Job that never finishes: every new run is skipped.
4. **`startingDeadlineSeconds`** too small: runs that start late are counted as missed.

```sh
kubectl -n wl-cronjob-fix patch cronjob cleanup -p '{"spec":{"schedule":"*/10 * * * *","suspend":false}}'
kubectl -n wl-cronjob-fix create job cleanup-manual --from=cronjob/cleanup
```

`create job --from=cronjob/...` copies the CronJob's Job template and marks the result with the
annotation `cronjob.kubernetes.io/instantiate: manual`. It's the standard way to test a CronJob
without waiting for its schedule.
