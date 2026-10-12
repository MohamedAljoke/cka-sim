kubectl --kubeconfig /opt/course/ar-kubeconfig-cert/auditor.kubeconfig config view --raw -o jsonpath='{.users[0].user.client-certificate-data}' | base64 -d > /tmp/auditor.crt
openssl x509 -in /tmp/auditor.crt -noout -subject -enddate
{
  echo "user: $(openssl x509 -in /tmp/auditor.crt -noout -subject -nameopt RFC2253 | grep -o 'CN=[^,]*' | cut -d= -f2)"
  echo "groups: $(openssl x509 -in /tmp/auditor.crt -noout -subject -nameopt RFC2253 | grep -o 'O=[^,]*' | cut -d= -f2 | paste -sd,)"
  echo "expires: $(openssl x509 -in /tmp/auditor.crt -noout -enddate | cut -d= -f2)"
} > /opt/course/ar-kubeconfig-cert/identity.txt
