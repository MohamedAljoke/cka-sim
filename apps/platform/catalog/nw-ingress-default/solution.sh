kubectl -n nw-ingress-default patch ingress site --type=merge \
  -p '{"spec":{"defaultBackend":{"service":{"name":"fallback","port":{"number":80}}}}}'
sleep 5
