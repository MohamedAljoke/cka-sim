{
  echo "user: $(kubectl auth whoami -o jsonpath='{.status.userInfo.username}')"
  echo "groups: $(kubectl auth whoami -o jsonpath='{.status.userInfo.groups[*]}' | tr ' ' ',')"
} > /opt/course/ar-whoami/me.txt
kubectl auth can-i --list -n ar-whoami --as=system:serviceaccount:ar-whoami:viewer > /opt/course/ar-whoami/viewer.txt
