kubectl -n nw-ingress-tls create secret tls secure-tls \
  --cert=/opt/course/nw-ingress-tls/tls.crt --key=/opt/course/nw-ingress-tls/tls.key
kubectl -n nw-ingress-tls patch ingress secure --type=merge \
  -p '{"spec":{"tls":[{"hosts":["secure.nw-ingress-tls.test"],"secretName":"secure-tls"}]}}'
sleep 5
