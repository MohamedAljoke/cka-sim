NS=wl-spread
tsc() { kubectl -n $NS get deployment web -o jsonpath="{.spec.template.spec.topologySpreadConstraints[0].$1}"; }
constraint() { eq 1 "$(tsc maxSkew)" && eq kubernetes.io/hostname "$(tsc topologyKey)" && eq DoNotSchedule "$(tsc whenUnsatisfiable)"; }
on() { kubectl -n $NS get pods -l app=web --field-selector spec.nodeName="$1",status.phase=Running -o name | wc -l; }
ready() { eq 4 "$(kubectl -n $NS get deployment web -o jsonpath='{.status.readyReplicas}')" && eq 4 "$(kubectl -n $NS get deployment web -o jsonpath='{.status.updatedReplicas}')"; }
even() { ready && eq 2 "$(on cka-sim-worker)" && eq 2 "$(on cka-sim-worker2)"; }
constraint && wait_for 120 even

check 2 "web has a hostname spread constraint with maxSkew 1 and DoNotSchedule" constraint
spread_ready() { constraint && ready; }
check 2 "All 4 Pods of web are Ready with the constraint in place" spread_ready
check 2 "2 Pods run on each worker node" even
