kubectl -n nw-gateway-header patch httproute app --type=json -p '[{"op":"add","path":"/spec/rules/0","value":{"matches":[{"headers":[{"name":"x-version","value":"v2"}]}],"backendRefs":[{"name":"v2","port":80}]}}]'
sleep 5
