An Ingress backend names a Service **in the Ingress's own Namespace**. There's no `namespace` field,
so `portal`'s `/billing` path pointed at a `billing` Service that doesn't exist there:

```bash
kubectl -n nw-ingress-cross describe ingress portal   # billing:80 (<error: services "billing" not found>)
```

Controllers merge every Ingress for the same host, so each team keeps its own Ingress next to its
Service:

```bash
kubectl -n nw-ingress-cross-billing create ingress billing --class=nginx \
  --rule='portal.nw-ingress-cross.test/billing*=billing:80'
kubectl -n nw-ingress-cross edit ingress portal     # delete the /billing path
```

Remove the old path so two Ingresses don't claim `/billing` on the same host. Which one wins would
then be up to the controller.

Other ways around it: a Service of type `ExternalName` in `nw-ingress-cross` pointing at
`billing.nw-ingress-cross-billing.svc.cluster.local` (some controllers refuse those), or Gateway API,
where an HTTPRoute can reference another Namespace's Service through a `ReferenceGrant`.
