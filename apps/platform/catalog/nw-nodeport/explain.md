A **NodePort** Service is a ClusterIP Service plus a port, from `30000-32767` by default, that
kube-proxy opens on **every** node. Traffic to any node's IP on that port reaches the Service,
even on a node where none of its Pods run.

```sh
kubectl -n nw-nodeport create service nodeport web --tcp=80:80 --node-port=30080
# or: kubectl -n nw-nodeport expose deployment web --type=NodePort --port=80, then set nodePort: 30080
kubectl -n nw-nodeport get svc web                 # PORT(S) 80:30080/TCP
kubectl get nodes -o wide                          # INTERNAL-IP
curl http://<node-ip>:30080
```

`kubectl create service nodeport` picks the selector `app=web` from the Service name, which matches
here because `create deployment web` labels its Pods `app=web`. `expose` can't set the nodePort
directly, so you would `kubectl edit` the Service afterwards.

Three ports, three jobs: `nodePort` on the node, `port` on the ClusterIP, `targetPort` in the
Pod. Leave `nodePort` out and Kubernetes picks a free one.
