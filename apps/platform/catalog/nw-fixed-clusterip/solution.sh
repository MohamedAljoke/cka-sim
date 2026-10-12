kubectl -n nw-fixed-clusterip create service clusterip cache --tcp=80:80 --clusterip=10.96.200.200
kubectl -n nw-fixed-clusterip create service clusterip cache-outside --tcp=80:80 --clusterip=10.200.0.10 2> /opt/course/nw-fixed-clusterip/error.txt || true
