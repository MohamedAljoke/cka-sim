kubectl -n nw-headless create service clusterip db --tcp=80:80 --clusterip=None
sleep 5
kubectl -n nw-headless exec client -- nslookup db.nw-headless.svc.cluster.local | sed -n '/^Name:/,$p' | awk '/^Address/{print $2}' > /opt/course/nw-headless/ips.txt
