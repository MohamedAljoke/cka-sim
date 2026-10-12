pki=/etc/kubernetes/pki
# openssl x509 -in <cert> -noout -issuer -> CN = kubernetes / etcd-ca / front-proxy-ca
for c in apiserver.crt apiserver-etcd-client.crt front-proxy-client.crt etcd/server.crt; do
  for ca in $pki/ca.crt $pki/etcd/ca.crt $pki/front-proxy-ca.crt; do
    if openssl verify -CAfile "$ca" "$pki/$c" >/dev/null 2>&1; then echo "$c: $ca"; fi
  done
done > /opt/course/ar-ca-map/ca.txt
