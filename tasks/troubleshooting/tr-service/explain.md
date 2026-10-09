**Concept.** A Service never talks to a Deployment. It selects **Pods by label**, and the
EndpointSlice controller keeps a list of the matching, Ready Pods' IPs. Traffic to the Service's
`port` is forwarded to each Pod's `targetPort`.

So there are two independent ways to break it, and this task had both:
1. **Selector matches nothing** → `k get endpoints web` is empty. Compare
   `k get svc web -o yaml` selector with `k get pods --show-labels`.
2. **Wrong targetPort** → endpoints exist, but connections are refused, because nothing listens
   on 8080 in the Pod.

**Fast test from inside the cluster:** `k run tmp --rm -it --image=busybox:1.36 --restart=Never --
wget -qO- -T 2 http://web.shop` — `web.shop` resolves through CoreDNS as `<service>.<namespace>`.
