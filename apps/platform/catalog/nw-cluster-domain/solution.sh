# grep -A1 -E 'clusterDomain|clusterDNS' /var/lib/kubelet/config.yaml
# kubectl run tmp --image=busybox:1.36 --restart=Never --rm -i -- cat /etc/resolv.conf
cfg=/var/lib/kubelet/config.yaml
{
  echo "clusterDomain: $(sed -n 's/^clusterDomain: *//p' $cfg)"
  echo "nameserver: $(grep -A1 '^clusterDNS:' $cfg | tail -1 | tr -d ' -')"
} > /opt/course/nw-cluster-domain/dns.txt
