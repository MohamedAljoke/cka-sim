`replicas: 2` in the file is the first problem: every apply sets it back, and the HPA corrects it
15 seconds later. Remove the field, so the file stops claiming it.

That alone isn't enough, and that's the trap. `kubectl apply` (client-side) remembers what the
file said last time, in the annotation `kubectl.kubernetes.io/last-applied-configuration`. A field
that was in the last applied file and is missing now gets **deleted**, and a Deployment without
`replicas` defaults to `1`. Try it: the next apply drops `web` to one Pod.

So update the memory together with the file:

```sh
sed -i '/^  replicas:/d' /opt/course/wl-hpa-fight/web.yaml
kubectl apply set-last-applied -f /opt/course/wl-hpa-fight/web.yaml
```

`set-last-applied` only rewrites the annotation; it doesn't change the Deployment.
`kubectl apply edit-last-applied deployment/web` does the same in an editor. From then on, applies
leave `spec.replicas` to the HPA. The Kubernetes docs describe this under "Migrating Deployments
and StatefulSets to horizontal autoscaling".

With server-side apply (`kubectl apply --server-side`) the field managers keep track instead, and
dropping `replicas` from the file simply hands that field over.
