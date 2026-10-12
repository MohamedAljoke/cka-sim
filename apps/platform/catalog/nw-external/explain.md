Normally the EndpointSlice controller fills a Service's endpoints from the Pods its selector
matches. A Service **without a selector** gets no automatic endpoints. You provide them yourself,
which lets a cluster DNS name point at anything with an IP: a database VM, a service in another
cluster, or a backend you're migrating.

```yaml
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: legacy-db-1
  labels:
    kubernetes.io/service-name: legacy-db     # this label ties the slice to the Service
addressType: IPv4
ports:
- name: http          # must match the Service port's name ("" if the Service port has none)
  port: 8080          # the real port on the backend
endpoints:
- addresses: ["10.244.2.9"]
```

The traps:
- Without the `kubernetes.io/service-name` label, kube-proxy never links the slice to the Service.
- The slice's port **name** must equal the Service port's name. The slice's port number is what
  traffic is sent to; the Service's `targetPort` is ignored without a selector.
- The old way, an `Endpoints` object with the same name as the Service, still works but is
  deprecated.

If the outside backend has a DNS name rather than an IP, use a Service of type `ExternalName`.
