NS=wl-job-limits
job() { kubectl -n $NS get job migrate -o jsonpath="$1" 2>/dev/null; }
limits() { eq "60 300" "$(job '{.spec.activeDeadlineSeconds} {.spec.ttlSecondsAfterFinished}')"; }
complete() { limits && eq 1 "$(job '{.status.succeeded}')" && kubectl -n $NS logs job/migrate | grep -q done; }
limits && wait_for 60 complete

check 2 "migrate has activeDeadlineSeconds 60" eq 60 "$(job '{.spec.activeDeadlineSeconds}')"
check 2 "migrate has ttlSecondsAfterFinished 300" eq 300 "$(job '{.spec.ttlSecondsAfterFinished}')"
check 1 "migrate completed" complete
