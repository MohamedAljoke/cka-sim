fresh_ns wl-taints
for node in $(kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name} {.spec.taints[*].key}{"\n"}{end}' | awk '/ dedicated( |$)/{print $1}'); do
  kubectl taint node "$node" dedicated- >/dev/null
done
