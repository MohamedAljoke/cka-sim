NS=tr-native-sidecar
init() { kubectl -n $NS get pod report -o jsonpath="{.spec.initContainers[?(@.name==\"log-shipper\")].$1}" 2>/dev/null; }
native() { eq "busybox:1.36 Always" "$(init image) $(init restartPolicy)"; }
running() { native && eq True "$(kubectl -n $NS get pod report -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
ships() { running && kubectl -n $NS logs report -c log-shipper --tail=20 | grep -q 'report line'; }
native && wait_for 60 ships

check 3 "log-shipper is an init container with restartPolicy Always" native
check 2 "kubectl logs report -c log-shipper shows the app's lines" ships
