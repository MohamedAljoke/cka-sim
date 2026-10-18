kubectl -n nw-gateway-allowed patch gateway shared-gw --type=json -p '[{"op":"add","path":"/spec/listeners/0/allowedRoutes","value":{"namespaces":{"from":"Selector","selector":{"matchLabels":{"shared-gateway-access":"true"}}}}}]'
sleep 5
