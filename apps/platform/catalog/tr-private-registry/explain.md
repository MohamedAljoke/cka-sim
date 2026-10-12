The Pod's events name the problem:

```text
Failed to pull image "cka-sim-registry:5000/team/web:1.0": ... authorization failed: no basic auth credentials
```

The kubelet pulls with whatever credentials the Pod lists in `imagePullSecrets`. A Pod gets them
from its own spec, or **from its ServiceAccount**: the ServiceAccount admission plugin copies the
ServiceAccount's `imagePullSecrets` into every new Pod that runs as it.

```sh
kubectl -n tr-private-registry create secret docker-registry registry-creds \
  --docker-server=cka-sim-registry:5000 --docker-username=candidate --docker-password=cka-registry
kubectl -n tr-private-registry patch serviceaccount default \
  -p '{"imagePullSecrets":[{"name":"registry-creds"}]}'
kubectl -n tr-private-registry rollout restart deployment web
```

The copy happens only when a Pod is created, so the stuck Pods keep failing until they're
replaced. `rollout restart` (or deleting them) does that.

`--docker-server` must match the registry part of the image name exactly, port included.
`kubectl get secret registry-creds -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d` shows
what the kubelet will send.
