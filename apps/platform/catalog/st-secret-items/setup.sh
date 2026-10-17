fresh_ns st-secret-items
kubectl -n st-secret-items create secret generic tls-bundle \
  --from-literal=tls.crt='-----CERT-----' --from-literal=tls.key='-----PRIVATE KEY-----' --from-literal=ca.crt='-----CA-----' >/dev/null
