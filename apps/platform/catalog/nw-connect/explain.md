Every Pod gets its own IP, and the CNI makes every Pod reachable from every other Pod, on any
node, without NAT. A Service adds a stable virtual IP and a DNS name in front of Pods that come
and go. A throwaway Pod is the standard way to test both from inside the cluster:

```sh
kubectl -n nw-connect get pod backend -o wide                       # the Pod IP
kubectl -n nw-connect run tmp --image=busybox:1.36 --restart=Never --rm -i --quiet -- \
  wget -qO- -T 3 http://10.244.1.7:8080
kubectl -n nw-connect run tmp --image=busybox:1.36 --restart=Never --rm -i --quiet -- \
  wget -qO- -T 3 http://backend                 # same namespace: the short name works
#   from another namespace: http://backend.nw-connect or backend.nw-connect.svc.cluster.local
```

`--rm -i` attaches, prints the output and deletes the Pod when it exits. `--quiet` keeps kubectl's
own "pod deleted" line out of a redirected file. Without `--restart=Never`, kubectl would create
a Pod that restarts forever. In a script, use `--attach` instead of `-i`: `-i` forwards your
stdin into the Pod and can swallow the rest of a script fed through stdin.

Testing the Pod IP and the Service name separately tells you which layer is broken. If the IP
works but the name doesn't, look at the Service, its endpoints or DNS. If the IP fails too, look
at the Pod or the CNI.
