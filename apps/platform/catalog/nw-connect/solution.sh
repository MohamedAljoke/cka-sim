ip=$(kubectl -n nw-connect get pod backend -o jsonpath='{.status.podIP}')
kubectl -n nw-connect run tmp-ip --image=busybox:1.36 --restart=Never --rm --attach --quiet --command -- wget -qO- -T 3 "http://$ip:8080" > /opt/course/nw-connect/by-ip.txt
kubectl -n nw-connect run tmp-name --image=busybox:1.36 --restart=Never --rm --attach --quiet --command -- wget -qO- -T 3 http://backend > /opt/course/nw-connect/by-name.txt
