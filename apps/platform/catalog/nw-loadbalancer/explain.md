A **LoadBalancer** Service builds on a NodePort Service, which builds on a ClusterIP Service. On top
of those it asks for an outside load balancer that sends traffic to the nodePort on the nodes.
Kubernetes doesn't create that load balancer itself. A **cloud controller manager** (AWS, GCP,
Azure), or MetalLB or cloud-provider-kind on bare metal and kind, watches LoadBalancer Services
and fills in `status.loadBalancer.ingress`.

```sh
kubectl -n nw-loadbalancer expose deployment shop --type=LoadBalancer --port=80
kubectl -n nw-loadbalancer get svc shop         # EXTERNAL-IP <pending>, PORT(S) 80:3xxxx/TCP
kubectl -n nw-loadbalancer describe svc shop    # no events from a load balancer controller
curl http://<node-ip>:<nodePort>                # works anyway
```

With no controller, `EXTERNAL-IP` stays `<pending>` forever. That's not an error in the Service.
Everything below the load balancer still works, so the nodePort and the ClusterIP serve traffic as
usual.

`spec.loadBalancerClass` picks which controller should handle a Service when there is more than
one, and `allocateLoadBalancerNodePorts: false` skips the nodePort for load balancers that route
straight to Pods.
