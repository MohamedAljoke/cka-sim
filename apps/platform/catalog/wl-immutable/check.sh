NS=wl-immutable
cm() { kubectl -n $NS get configmap "$1" -o jsonpath="{.immutable} {.data.color}"; }
v1_ok() { eq "true green" "$(cm app-config-v1)"; }
v2_ok() { eq "true blue" "$(cm app-config-v2)"; }
ref() { kubectl -n $NS get deployment app -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="COLOR")].valueFrom.configMapKeyRef.name}'; }
D() { kubectl -n $NS get deployment app -o jsonpath="$1"; }
rolled() { eq app-config-v2 "$(ref)" && eq 2 "$(D '{.status.updatedReplicas}')" && eq 2 "$(D '{.status.readyReplicas}')"; }
blue() { rolled && eq blue "$(kubectl -n $NS exec deploy/app -- printenv COLOR)"; }
eq app-config-v2 "$(ref)" && wait_for 90 rolled

check 1 "app-config-v1 is immutable and still says green" v1_ok
check 2 "app-config-v2 is immutable and says blue" v2_ok
check 2 "Deployment app uses app-config-v2 and has rolled out" rolled
check 1 "The running Pods see COLOR=blue" blue
