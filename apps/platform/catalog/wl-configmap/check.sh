NS=wl-configmap
cm() { kubectl -n $NS get configmap app-config -o jsonpath="$1"; }
cm_ok() { eq debug "$(cm '{.data.LOG_LEVEL}')" && eq "$(cat "$COURSE/app.properties")" "$(cm '{.data.app\.properties}')"; }
running() { eq True "$(kubectl -n $NS get pod app -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
in_pod() { kubectl -n $NS exec app -- "$@"; }
file_ok() { running && eq "$(cat "$COURSE/app.properties")" "$(in_pod cat /etc/app/app.properties)" && eq app.properties "$(in_pod ls /etc/app | grep -v '^\.\.' | xargs)"; }
env_ok() { running && eq debug "$(in_pod printenv LOG_LEVEL)"; }
cm_ok && wait_for 60 running

check 2 "ConfigMap app-config has LOG_LEVEL=debug and the file as app.properties" cm_ok
check 2 "Pod app sees only app.properties in /etc/app" file_ok
check 2 "Pod app has LOG_LEVEL=debug in its environment" env_ok
