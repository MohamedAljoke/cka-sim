kubectl -n nw-ingress-pathtype patch ingress docs --type=json \
  -p '[{"op":"replace","path":"/spec/rules/0/http/paths/0/pathType","value":"Prefix"}]'
sleep 5
