NS=wl-vpa
vpa() { kubectl -n $NS get vpa api -o jsonpath="$1"; }
target() { vpa '{.status.recommendation.containerRecommendations[?(@.containerName=="api")].target.cpu} {.status.recommendation.containerRecommendations[?(@.containerName=="api")].target.memory}'; }
answer() { tr -d ' \r' < "$COURSE/target.txt" 2>/dev/null | sed 's/memory=/ memory=/' | head -1; }
requests() { kubectl -n $NS get pods -l app=api -o jsonpath='{range .items[*]}{.spec.containers[0].resources.requests.cpu}/{.spec.containers[0].resources.requests.memory} {end}' | xargs; }
off() { eq "Deployment api Off" "$(vpa '{.spec.targetRef.kind} {.spec.targetRef.name} {.spec.updatePolicy.updateMode}')" && eq "100m/64Mi 100m/64Mi" "$(requests)"; }
written() { set -- $(target); [ -n "$1" ] && eq "cpu=$1 memory=$2" "$(answer)"; }

check 2 "VPA api targets Deployment api with updateMode Off, and the Pods keep their requests" off
check 1 "VPA api shows a recommendation" eq True "$(vpa '{.status.conditions[?(@.type=="RecommendationProvided")].status}')"
check 2 "target.txt holds the recommended target CPU and memory for container api" written
