kubectl -n wl-job-limits create job migrate --image=busybox:1.36 --dry-run=client -o yaml -- sh -c 'echo migrating; sleep 5; echo done' |
  sed 's/^spec:$/spec:\n  activeDeadlineSeconds: 60\n  ttlSecondsAfterFinished: 300/' | kubectl apply -f -
kubectl -n wl-job-limits wait --for=condition=Complete job/migrate --timeout=90s
