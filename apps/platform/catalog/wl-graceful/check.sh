NS=wl-graceful
D() { kubectl -n $NS get deployment web -o jsonpath="$1"; }
prestop() { eq '["sleep","10"]' "$(D '{.spec.template.spec.containers[0].lifecycle.preStop.exec.command}')" || eq '["sh","-c","sleep 10"]' "$(D '{.spec.template.spec.containers[0].lifecycle.preStop.exec.command}')" || eq 10 "$(D '{.spec.template.spec.containers[0].lifecycle.preStop.sleep.seconds}')"; }
grace() { eq 45 "$(D '{.spec.template.spec.terminationGracePeriodSeconds}')"; }
rolled() { prestop && grace && eq "2 2" "$(D '{.status.updatedReplicas} {.status.readyReplicas}')" && [ -z "$(D '{.status.unavailableReplicas}')" ]; }
prestop && wait_for 120 rolled

check 2 "web's container sleeps 10 seconds in a preStop hook" prestop
check 2 "web's Pods have terminationGracePeriodSeconds 45" grace
check 1 "The change is rolled out" rolled
