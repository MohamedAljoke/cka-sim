# kubectl -n wl-cronjob-fix get cronjob cleanup -> SCHEDULE 0 0 31 2 * (Feb 31st never comes), SUSPEND True
kubectl -n wl-cronjob-fix patch cronjob cleanup -p '{"spec":{"schedule":"*/10 * * * *","suspend":false}}'
kubectl -n wl-cronjob-fix create job cleanup-manual --from=cronjob/cleanup
kubectl -n wl-cronjob-fix wait --for=condition=Complete job/cleanup-manual --timeout=60s
