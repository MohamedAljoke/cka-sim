echo api.nw-dns-names.svc.cluster.local > /opt/course/nw-dns-names/service.txt
ip=$(kubectl -n nw-dns-names get pod worker -o jsonpath='{.status.podIP}')
echo "${ip//./-}.nw-dns-names.pod.cluster.local" > /opt/course/nw-dns-names/pod.txt
kubectl -n nw-dns-names exec worker -- nslookup api.nw-dns-names.svc.cluster.local
kubectl -n nw-dns-names exec worker -- nslookup "${ip//./-}.nw-dns-names.pod.cluster.local"
