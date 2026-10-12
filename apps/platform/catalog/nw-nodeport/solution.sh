kubectl -n nw-nodeport create service nodeport web --tcp=80:80 --node-port=30080
kubectl get nodes -o wide
curl -s "http://$(kubectl get node cka-sim-worker2 -o jsonpath='{.status.addresses[0].address}'):30080" | head -4
