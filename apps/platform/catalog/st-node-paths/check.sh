answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/paths.txt" 2>/dev/null | head -1 | tr -d ' \r' | sed 's|/$||'; }
uid() { kubectl -n st-node-paths get pod cache -o jsonpath='{.metadata.uid}'; }
dir_ok() { eq "/var/lib/kubelet/pods/$(uid)/volumes/kubernetes.io~empty-dir/scratch" "$(answer emptydir)" && [ -f "$(answer emptydir)/state" ]; }
pid() { crictl inspect "$(crictl ps -q --name '^cache$' --namespace st-node-paths | head -1)" | jq -r '.info.pid'; }

check 2 "emptydir is the volume's directory on the node" dir_ok
check 2 "pid is the container's main process on the host" eq "$(pid)" "$(answer pid)"
