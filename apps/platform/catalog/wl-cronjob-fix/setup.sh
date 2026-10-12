fresh_ns wl-cronjob-fix
kubectl -n wl-cronjob-fix create cronjob cleanup --image=busybox:1.36 --schedule='0 0 31 2 *' -- sh -c 'echo cleaning up' >/dev/null
kubectl -n wl-cronjob-fix patch cronjob cleanup -p '{"spec":{"suspend":true}}' >/dev/null
