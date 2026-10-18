for n in cka-sim-worker cka-sim-worker2; do
  ip=$(kubectl get node $n -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}')
  curl -s -m 3 "http://$ip:30090" >/dev/null && echo $n
done > /opt/course/nw-etp-local/answering.txt
kubectl -n nw-etp-local patch service web -p '{"spec":{"externalTrafficPolicy":"Cluster"}}'
