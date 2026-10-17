answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/dns.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
cm() { kubectl -n kube-system get deployment coredns -o jsonpath='{.spec.template.spec.volumes[?(@.configMap)].configMap.name}'; }
upstream() { kubectl -n kube-system get configmap "$(cm)" -o jsonpath='{.data.Corefile}' | awk '$1=="forward"{print $3; exit}'; }

check 1 "service-ip is the kube-dns ClusterIP" eq "$(kubectl -n kube-system get service kube-dns -o jsonpath='{.spec.clusterIP}')" "$(answer service-ip)"
check 1 "configmap is kube-system/coredns" eq "kube-system/$(cm)" "$(answer configmap)"
check 1 "upstream is the forward target" eq "$(upstream)" "$(answer upstream)"
