NS=nw-dns-config
answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/client.txt" 2>/dev/null | head -1 | tr -d '\r' | xargs; }
conf() { kubectl -n $NS exec "$1" -- cat /etc/resolv.conf 2>/dev/null; }
search() { conf client | sed -n 's/^search //p' | xargs; }
ndots() { conf client | grep -o 'ndots:[0-9]*' | cut -d: -f2; }
custom_ok() {
  eq None "$(kubectl -n $NS get pod custom-dns -o jsonpath='{.spec.dnsPolicy}' 2>/dev/null)" &&
    eq "nameserver 10.96.0.10" "$(conf custom-dns | grep ^nameserver | xargs)" &&
    eq "search nw-dns-config.svc.cluster.local" "$(conf custom-dns | grep ^search | xargs)" &&
    conf custom-dns | grep -q 'ndots:2'
}
kubectl -n $NS get pod custom-dns >/dev/null 2>&1 && wait_for 60 custom_ok

check 1 "search holds client's search domains" eq "$(search)" "$(answer search)"
check 1 "ndots holds client's ndots" eq "$(ndots)" "$(answer ndots)"
check 3 "custom-dns uses dnsPolicy None with the given nameserver, search and ndots" custom_ok
