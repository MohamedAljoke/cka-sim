kubectl -n nw-gateway-status get httproute inventory \
  -o jsonpath='{range .status.parents[0].conditions[?(@.status=="False")]}{.type} {.reason}{"\n"}{end}' \
  > /opt/course/nw-gateway-status/condition.txt
kubectl -n nw-gateway-status patch httproute inventory --type=json -p '[{"op":"replace","path":"/spec/rules/0/backendRefs/0/name","value":"inventory"}]'
sleep 5
