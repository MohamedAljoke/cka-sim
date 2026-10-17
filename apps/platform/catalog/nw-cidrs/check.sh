answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/cidrs.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
flag() { grep -o -- "--$2=[^ ]*" "/etc/kubernetes/manifests/$1.yaml" | head -1 | cut -d= -f2; }

check 1 "pods is the controller-manager's --cluster-cidr" eq "$(flag kube-controller-manager cluster-cidr)" "$(answer pods)"
check 2 "services is the API server's --service-cluster-ip-range" eq "$(flag kube-apiserver service-cluster-ip-range)" "$(answer services)"
check 1 "cka-sim-worker is that node's spec.podCIDR" eq "$(kubectl get node cka-sim-worker -o jsonpath='{.spec.podCIDR}')" "$(answer cka-sim-worker)"
