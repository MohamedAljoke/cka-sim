Generators create ConfigMaps and Secrets from files or literals, and give each one a **name
suffix made from a hash of its content** (`app-config-7b8m2c6tdk`). Kustomize rewrites every
reference in the Deployment to that name. Change the content and the name changes, so the Pod
template changes, so the Deployment rolls out new Pods. Config changes become rollouts, with
history and rollback.

```yaml
configMapGenerator:
- name: app-config
  files:
  - app.properties          # key app.properties, value the file
secretGenerator:
- name: app-secret
  literals:
  - password=s3cr3t
```

```sh
kubectl kustomize /opt/course/ar-kustomize-gen     # see the generated names before applying
kubectl apply -k /opt/course/ar-kustomize-gen
kubectl -n ar-kustomize-gen get cm,secret          # the old app-config-<hash> stays until pruned
```

Compare with a plain ConfigMap: editing it updates the mounted file in running Pods after a
minute or so, but env vars never change, and nothing records that the config changed. The
hash suffix trades that for clean rollouts. Turn it off with `generatorOptions:
disableNameSuffixHash: true` only when something outside Kustomize needs a fixed name.
