A Service's `targetPort` can be a **name** instead of a number. It then points at whatever
container port has that name in each Pod:

```yaml
# Service
ports:
- port: 80
  targetPort: http
# Pod template
ports:
- name: http
  containerPort: 9090
```

The app's port becomes the Pod's own business: change it in the Deployment and roll out, and the
Service follows with no edit. During the rollout, old Pods (8080) and new Pods (9090) both receive
traffic on their own port, because each endpoint is resolved per Pod. With a numeric `targetPort`,
the same change would break the Service until someone updated it too.

```sh
kubectl -n nw-named-port get endpointslices -o yaml | grep -A3 ports:     # the resolved port number
```

Port names must be lowercase, at most 15 characters, and unique within a Pod. Named ports also
work in NetworkPolicies and in probes (`port: http`).
