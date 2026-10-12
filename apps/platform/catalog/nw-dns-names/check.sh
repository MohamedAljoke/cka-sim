NS=nw-dns-names
domain() { sed -n 's/^clusterDomain: *//p' /var/lib/kubelet/config.yaml; }
answer() { tr -d ' \r' < "$COURSE/$1" 2>/dev/null | grep . | head -1 | sed 's/\.$//'; }
pod_name() { echo "$(kubectl -n $NS get pod worker -o jsonpath='{.status.podIP}' | tr . -).$NS.pod.$(domain)"; }

check 2 "service.txt holds api.nw-dns-names.svc.<cluster domain>" eq "api.$NS.svc.$(domain)" "$(answer service.txt)"
check 2 "pod.txt holds the IP-based Pod name" eq "$(pod_name)" "$(answer pod.txt)"
