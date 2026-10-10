# A forgotten taint would keep other tasks' Pods off the node.
tainted=$(kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name} {.spec.taints[*].key}{"\n"}{end}' | awk '/ dedicated( |$)/{print $1}')
[ -z "$tainted" ] && exit 0
for node in $tainted; do
  kubectl taint node "$node" dedicated- >/dev/null
done
