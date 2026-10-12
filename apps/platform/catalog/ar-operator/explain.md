An operator is two things: a **CRD** that adds a new kind to the API, and a **controller** that watches
objects of that kind and makes the cluster match them.

```bash
kubectl apply -f /opt/course/ar-operator/web-operator.yaml
kubectl get crd webapps.cka-sim.io
kubectl -n web-operator rollout status deployment web-operator
kubectl explain webapp.spec                   # the fields the CRD's schema defines
```

```yaml
apiVersion: cka-sim.io/v1
kind: WebApp
metadata:
  name: shop
  namespace: ar-operator
spec:
  image: nginx:1.27-alpine
  replicas: 2
  port: 80
```

```bash
kubectl -n ar-operator get webapp,deploy,svc     # WebApp shop, plus the Deployment and Service it built
kubectl -n ar-operator get deploy shop -o jsonpath='{.metadata.ownerReferences}'
kubectl -n web-operator logs deploy/web-operator
```

- **The operator owns what it builds** through `ownerReferences`. Delete the `WebApp` and the garbage
  collector removes its Deployment and Service.
- **It reconciles:** scale the Deployment by hand and it's set back to `spec.replicas`. Change the
  `WebApp` instead.
- **Order matters:** a custom resource can't be created before its CRD is `Established`.

Real operators (cert-manager, Prometheus Operator, CloudNativePG) work the same way, and are often
installed with Helm (`helm install ... --create-namespace`).
