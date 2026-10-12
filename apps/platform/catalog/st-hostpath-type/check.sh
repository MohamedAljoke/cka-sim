NS=st-hostpath-type
P() { kubectl -n $NS get pod agent -o jsonpath="$1" 2>/dev/null; }
typed() { eq "/var/lib/agent-data DirectoryOrCreate" "$(P '{.spec.volumes[?(@.name=="state")].hostPath.path} {.spec.volumes[?(@.name=="state")].hostPath.type}')"; }
runs() { typed && eq Running "$(P '{.status.phase}')"; }
written() { runs && [ -s /var/lib/agent-data/starts ]; }
typed && wait_for 60 runs

check 2 "agent's hostPath keeps the path and uses type DirectoryOrCreate" typed
check 2 "Pod agent runs" runs
check 1 "The agent wrote its state to /var/lib/agent-data on the node" written
