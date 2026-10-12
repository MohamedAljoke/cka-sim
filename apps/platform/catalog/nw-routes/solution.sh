cidr=$(kubectl get node cka-sim-worker2 -o jsonpath='{.spec.podCIDR}')
# ip route | grep "$cidr" -> 10.244.2.0/24 via 172.18.0.4 dev eth0
r=$(ip route | grep "^$cidr ")
{
  echo "podCIDR: $cidr"
  echo "via: $(echo "$r" | grep -o 'via [^ ]*' | cut -d' ' -f2)"
  echo "dev: $(echo "$r" | grep -o 'dev [^ ]*' | cut -d' ' -f2)"
} > /opt/course/nw-routes/route.txt
