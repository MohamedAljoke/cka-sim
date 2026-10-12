kubectl -n wl-jobs create job batch --image=busybox:1.36 --dry-run=client -o yaml -- echo processed |
  sed 's/^spec:$/spec:\n  completions: 6\n  parallelism: 2\n  backoffLimit: 3/' | kubectl apply -f -
kubectl -n wl-jobs create cronjob report --image=busybox:1.36 --schedule='*/15 * * * *' --dry-run=client -o yaml -- date |
  sed 's/^spec:$/spec:\n  successfulJobsHistoryLimit: 2\n  failedJobsHistoryLimit: 1\n  concurrencyPolicy: Forbid/' | kubectl apply -f -
kubectl -n wl-jobs wait --for=condition=Complete job/batch --timeout=120s
