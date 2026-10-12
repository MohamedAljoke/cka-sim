# kubectl -n kube-system get cm kube-proxy -o yaml | grep mode -> mode: iptables
chain=$(iptables-save -t nat | grep 'nw-kube-proxy-mode/api' | grep -o 'KUBE-SVC-[A-Z0-9]*' | head -1)
{
  echo "mode: $(kubectl -n kube-system get cm kube-proxy -o jsonpath='{.data.config\.conf}' | awk '/^mode:/{gsub(/"/,"",$2); print $2}')"
  echo "chain: $chain"
  echo "endpoints: $(iptables-save -t nat | grep "^-A $chain " | grep -o 'KUBE-SEP-[A-Z0-9]*' | sort -u | wc -l)"
} > /opt/course/nw-kube-proxy-mode/proxy.txt
