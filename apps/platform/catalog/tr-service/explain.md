A Service never talks to a Deployment. It selects **Pods by label**, and the EndpointSlice
controller keeps a list of the matching, Ready Pods' IPs. Traffic to the Service's `port` is
forwarded to each Pod's `targetPort`.

So there are two independent ways to break it, and this task had both:

1. **The selector matches nothing**, so the Service has no endpoints. Compare the selector in
   `kubectl -n shop get service web -o yaml` with `kubectl -n shop get pods --show-labels`.
2. **The targetPort is wrong**, so endpoints exist but connections are refused, because nothing
   listens on 8080 in the Pod.

```sh
kubectl -n shop patch service web --type=json -p '[
  {"op":"replace","path":"/spec/selector","value":{"app":"web"}},
  {"op":"replace","path":"/spec/ports/0/targetPort","value":80}]'
```

`kubectl -n shop edit service web` works too. Test it from inside the cluster:

```sh
kubectl run tmp --rm -it --image=busybox:1.36 --restart=Never -- wget -qO- -T 2 http://web.shop
```

`web.shop` resolves through CoreDNS as `<service>.<namespace>`.
