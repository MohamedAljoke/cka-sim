NS=wl-jobs
job() { kubectl -n $NS get job batch -o jsonpath="$1"; }
job_spec() { eq "6 2 3" "$(job '{.spec.completions} {.spec.parallelism} {.spec.backoffLimit}')"; }
complete() { job_spec && eq 6 "$(job '{.status.succeeded}')"; }
cj() { kubectl -n $NS get cronjob report -o jsonpath="$1"; }
schedule_ok() { [[ "$(cj '{.spec.schedule}')" =~ ^(\*/15|0,15,30,45)\ \*\ \*\ \*\ \*$ ]]; }
history_ok() { eq "2 1 Forbid" "$(cj '{.spec.successfulJobsHistoryLimit} {.spec.failedJobsHistoryLimit} {.spec.concurrencyPolicy}')"; }
job_spec && wait_for 90 complete

check 1 "Job batch has completions 6, parallelism 2, backoffLimit 3" job_spec
check 2 "Job batch completed 6 times" complete
check 1 "CronJob report runs every 15 minutes" schedule_ok
check 2 "CronJob report keeps 2 successful and 1 failed Job and forbids overlaps" history_ok
