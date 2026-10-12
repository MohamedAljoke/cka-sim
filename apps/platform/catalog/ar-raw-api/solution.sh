kubectl get --raw /api/v1/namespaces/ar-raw-api/pods > /opt/course/ar-raw-api/pods.json
kubectl get --raw /apis/apps/v1/namespaces/ar-raw-api/deployments/web | jq .spec.replicas > /opt/course/ar-raw-api/replicas.txt
