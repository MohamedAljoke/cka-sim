NS=wl-probes
c() { kubectl -n $NS get deployment api -o jsonpath="{.spec.template.spec.containers[0].$1}"; }
startup() { eq "/ 80 2 30" "$(c 'startupProbe.httpGet.path') $(c 'startupProbe.httpGet.port') $(c 'startupProbe.periodSeconds') $(c 'startupProbe.failureThreshold')"; }
liveness() { eq "80 10" "$(c 'livenessProbe.tcpSocket.port') $(c 'livenessProbe.periodSeconds')"; }
readiness() { eq '["cat","/usr/share/nginx/html/index.html"] 5' "$(c 'readinessProbe.exec.command') $(c 'readinessProbe.periodSeconds')"; }
D() { kubectl -n $NS get deployment api -o jsonpath="$1"; }
ready() { startup && liveness && readiness && eq 2 "$(D '{.status.updatedReplicas}')" && eq 2 "$(D '{.status.readyReplicas}')"; }
startup && wait_for 120 ready

check 2 "api has an HTTP startup probe on / port 80, period 2, failureThreshold 30" startup
check 2 "api has a TCP liveness probe on port 80, period 10" liveness
check 2 "api has an exec readiness probe that cats index.html, period 5" readiness
check 1 "Both Pods of api are Ready with the probes" ready
