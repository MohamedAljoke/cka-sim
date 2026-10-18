helm -n ar-helm-history get metadata api --revision 2 | awk -F': ' '$1 == "VERSION" { print $2 }' > /opt/course/ar-helm-history/version.txt
helm -n ar-helm-history get values api --revision 2 -o yaml > /opt/course/ar-helm-history/values.yaml
