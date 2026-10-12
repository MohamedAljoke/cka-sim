```bash
kubectl annotate ingressclass nginx ingressclass.kubernetes.io/is-default-class=true
kubectl get ingressclass                      # nginx ... (default)
kubectl delete -f /opt/course/nw-ingress-class/legacy.yaml
kubectl apply -f /opt/course/nw-ingress-class/legacy.yaml
kubectl -n nw-ingress-class get ingress legacy   # CLASS nginx
```

- **The default class is filled in when the object is created,** by the `DefaultIngressClass`
  admission plugin. An Ingress created earlier without a class keeps none, so you must create it
  again (or set the class on it).
- **Only one IngressClass should carry the annotation.** With two, new Ingresses without a class are
  rejected.
- **The older `kubernetes.io/ingress.class` annotation** on an Ingress is deprecated. Use
  `spec.ingressClassName`, or rely on the default.

Some controllers can also be told to serve Ingresses that have no class at all (ingress-nginx's
`--watch-ingress-without-class`). The portable fix is the default IngressClass.
