answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/dns.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
cfg=/var/lib/kubelet/config.yaml

check 1 "clusterDomain is right" eq "$(sed -n 's/^clusterDomain: *//p' $cfg)" "$(answer clusterDomain)"
check 1 "nameserver is the cluster DNS IP" eq "$(grep -A1 '^clusterDNS:' $cfg | tail -1 | tr -d ' -')" "$(answer nameserver)"
