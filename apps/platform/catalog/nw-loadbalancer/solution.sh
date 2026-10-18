kubectl -n nw-loadbalancer expose deployment shop --type=LoadBalancer --port=80
kubectl -n nw-loadbalancer get service shop --no-headers | awk '{print $4}' > /opt/course/nw-loadbalancer/external-ip.txt
kubectl -n nw-loadbalancer get service shop -o jsonpath='{.spec.ports[0].nodePort}' > /opt/course/nw-loadbalancer/node-port.txt
