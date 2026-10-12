kubectl -n nw-gateway-canary patch httproute shop --type=json -p '[{"op":"replace","path":"/spec/rules/0/backendRefs","value":[{"name":"stable","port":80,"weight":80},{"name":"canary","port":80,"weight":20}]}]'
sleep 5
