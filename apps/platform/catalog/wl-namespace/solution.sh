kubectl create namespace galaxy
kubectl label namespace galaxy env=training
kubectl -n galaxy run probe --image=nginx:1.27
kubectl -n galaxy create deployment web --image=nginx:1.27 --replicas=2
# kubectl get pods -A | grep hermes
kubectl get pods -A --field-selector metadata.name=hermes -o jsonpath='{.items[0].metadata.namespace}' > /opt/course/wl-namespace/hermes.txt
kubectl -n galaxy wait pod probe --for=condition=Ready --timeout=120s
kubectl -n galaxy rollout status deployment/web --timeout=120s
