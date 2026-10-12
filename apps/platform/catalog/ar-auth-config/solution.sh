m=/etc/kubernetes/manifests/kube-apiserver.yaml
# grep -E 'client-ca|anonymous|authorization-mode' $m -> no --anonymous-auth flag: the default (true) applies
{
  echo "client-ca-file: $(grep -o -- '--client-ca-file=[^ ]*' $m | cut -d= -f2)"
  echo "anonymous-auth: true"
  echo "authorization-mode: $(grep -o -- '--authorization-mode=[^ ]*' $m | cut -d= -f2)"
} > /opt/course/ar-auth-config/auth.txt
