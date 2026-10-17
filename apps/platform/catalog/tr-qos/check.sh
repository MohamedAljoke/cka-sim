NS=tr-qos
answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/classes.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
listed() { eq Burstable "$(answer payments)" && eq Guaranteed "$(answer search)" && eq BestEffort "$(answer nightly)"; }
P() { kubectl -n $NS get pod payments -o jsonpath="$1" 2>/dev/null; }
guaranteed() { eq "Guaranteed Running" "$(P '{.status.qosClass} {.status.phase}')"; }
values() { guaranteed && eq "200m 128Mi 200m 128Mi" "$(P '{.spec.containers[0].resources.requests.cpu} {.spec.containers[0].resources.requests.memory} {.spec.containers[0].resources.limits.cpu} {.spec.containers[0].resources.limits.memory}')"; }
in_file() { grep -q 'limits:' "$COURSE/payments.yaml" && grep -q '128Mi' "$COURSE/payments.yaml"; }
in_file && wait_for 60 guaranteed

check 2 "classes.txt gives the original classes: Burstable, Guaranteed, BestEffort" listed
check 2 "Pod payments runs with QoS class Guaranteed" guaranteed
from_file() { values && in_file; }
check 1 "It requests and is limited to 200m CPU and 128Mi memory, from payments.yaml" from_file
