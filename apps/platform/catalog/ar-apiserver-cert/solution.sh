crt=/etc/kubernetes/pki/apiserver.crt
openssl x509 -in $crt -noout -text | grep -A1 'Subject Alternative Name'
names=$(openssl x509 -in $crt -noout -ext subjectAltName | grep -o 'DNS:[^,]*' | cut -d: -f2 | paste -sd,)
{
  echo "issuer: $(openssl x509 -in $crt -noout -issuer | sed -n 's/.*CN *= *\([^,]*\).*/\1/p')"
  echo "expires: $(openssl x509 -in $crt -noout -enddate | cut -d= -f2)"
  echo "dns-names: $names"
  echo "k8s.cka.internal: $(echo "$names" | tr ',' '\n' | grep -qx k8s.cka.internal && echo yes || echo no)"
} > /opt/course/ar-apiserver-cert/cert.txt
