# kubectl -n tr-crashloop get pods                   -> api-... CrashLoopBackOff
# kubectl -n tr-crashloop logs deploy/api --previous -> FATAL: /etc/api/config.yaml not found
# kubectl -n tr-crashloop get configmap api-config -o yaml -> the key is config.yml
{ kubectl -n tr-crashloop logs deploy/api --previous; kubectl -n tr-crashloop logs deploy/api; } 2>/dev/null | grep -m1 FATAL > /opt/course/tr-crashloop/error.txt
kubectl -n tr-crashloop get configmap api-config -o jsonpath='{.data.config\.yml}' > /tmp/config.yaml
kubectl -n tr-crashloop create configmap api-config --from-file=config.yaml=/tmp/config.yaml --dry-run=client -o yaml | kubectl replace -f -
kubectl -n tr-crashloop rollout restart deployment api
kubectl -n tr-crashloop rollout status deployment api --timeout=90s
