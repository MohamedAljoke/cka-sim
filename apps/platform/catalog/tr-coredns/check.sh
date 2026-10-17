corefile() { kubectl -n kube-system get configmap coredns -o jsonpath='{.data.Corefile}'; }
plugin_ok() { corefile | grep -q '^    kubernetes cluster.local in-addr.arpa ip6.arpa {$'; }
kept() { plugin_ok && corefile | grep -q '^    forward . /etc/resolv.conf' && corefile | grep -q '^    cache 30'; }
ready() { eq 2 "$(kubectl -n kube-system get deployment coredns -o jsonpath='{.status.readyReplicas}')"; }
resolves() { kubectl -n tr-coredns exec client -- nslookup -timeout=2 kubernetes.default.svc.cluster.local | grep -q 'Address:.*10\.96\.0\.1'; }
if plugin_ok; then
  wait_for 120 ready
  wait_for 30 resolves
fi

check 2 "The Corefile loads the kubernetes plugin again" plugin_ok
check 2 "Both CoreDNS Pods are Ready" ready
check 3 "Pod client resolves kubernetes.default.svc.cluster.local" resolves
check 1 "The rest of the Corefile is unchanged" kept
