NS=wl-cronjob-fix
cj() { kubectl -n $NS get cronjob cleanup -o jsonpath="$1"; }
schedule_ok() { [[ "$(cj '{.spec.schedule}')" =~ ^(\*/10|0,10,20,30,40,50|0-59/10)\ \*\ \*\ \*\ \*$ ]]; }
active() { schedule_ok && [ "$(cj '{.spec.suspend}')" != true ]; }
job() { kubectl -n $NS get job cleanup-manual -o jsonpath="$1" 2>/dev/null; }
from_cronjob() { eq "manual cleanup" "$(job '{.metadata.annotations.cronjob\.kubernetes\.io/instantiate} {.metadata.ownerReferences[0].name}')"; }
complete() { from_cronjob && eq 1 "$(job '{.status.succeeded}')"; }
from_cronjob && wait_for 60 complete

check 2 "cleanup is scheduled every 10 minutes" schedule_ok
check 2 "cleanup is not suspended" active
check 2 "Job cleanup-manual was created from the CronJob and completed" complete
