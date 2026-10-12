fresh_ns tr-besteffort-a
fresh_ns tr-besteffort-b
fresh_course
pod() { kubectl -n "$1" run "$2" --image=busybox:1.36 ${3:+--overrides="$3"} --command -- sleep 1d >/dev/null; }
res='{"spec":{"containers":[{"name":"%s","image":"busybox:1.36","command":["sleep","1d"],"resources":%s}]}}'
pod tr-besteffort-a frontend
pod tr-besteffort-a backend "$(printf "$res" backend '{"requests":{"memory":"32Mi"}}')"
pod tr-besteffort-a cache
pod tr-besteffort-b worker "$(printf "$res" worker '{"requests":{"cpu":"50m","memory":"32Mi"},"limits":{"cpu":"50m","memory":"32Mi"}}')"
pod tr-besteffort-b cron-runner
pod tr-besteffort-b metrics "$(printf "$res" metrics '{"limits":{"cpu":"100m"}}')"
kubectl -n tr-besteffort-a wait --for=condition=Ready pod --all --timeout=120s >/dev/null
kubectl -n tr-besteffort-b wait --for=condition=Ready pod --all --timeout=120s >/dev/null
