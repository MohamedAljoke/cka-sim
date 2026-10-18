# kubectl -n kube-system get deploy coredns -o yaml | grep -A3 configMap -> name: coredns
{
  echo "service-ip: $(kubectl -n kube-system get svc kube-dns -o jsonpath='{.spec.clusterIP}')"
  echo "configmap: kube-system/coredns"
  echo "upstream: $(kubectl -n kube-system get cm coredns -o jsonpath='{.data.Corefile}' | awk '$1=="forward"{print $3; exit}')"
} > /opt/course/nw-coredns-facts/dns.txt
