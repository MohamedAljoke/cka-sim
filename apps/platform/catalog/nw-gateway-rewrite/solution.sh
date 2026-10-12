kubectl -n nw-gateway-rewrite patch httproute legacy --type=json -p '[{"op":"add","path":"/spec/rules/0/filters","value":[{"type":"URLRewrite","urlRewrite":{"path":{"type":"ReplacePrefixMatch","replacePrefixMatch":"/"}}}]}]'
sleep 5
