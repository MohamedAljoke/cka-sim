kubectl -n nw-netpol-and-or patch networkpolicy api-from-monitoring --type=json -p '[{"op":"replace","path":"/spec/ingress/0/from","value":[{"namespaceSelector":{"matchLabels":{"team":"monitoring"}},"podSelector":{"matchLabels":{"app":"prometheus"}}}]}]'
sleep 5
