Kustomize, built into kubectl, customises plain YAML without templates. A **base** holds the
shared manifests, and an **overlay** points at the base and lists changes. The base stays untouched,
so dev, staging and prod overlays can share it.

```yaml
# overlays/prod/kustomization.yaml
resources:
- ../../base
namespace: ar-kustomize
replicas:
- name: shop
  count: 3
images:
- name: nginx              # matches by image name, whatever the tag
  newTag: 1.27-alpine
labels:
- pairs:
    env: prod
  includeTemplates: true   # also label the Pod template
```

```sh
kubectl kustomize /opt/course/ar-kustomize/overlays/prod      # render only: check before applying
kubectl apply -k /opt/course/ar-kustomize/overlays/prod
```

Labels: `labels` with `includeTemplates: true` labels the object and its Pod template but leaves
the selector alone. The older `commonLabels` also changes `spec.selector`, which is immutable on
an existing Deployment and fails the apply. Use `includeSelectors` only for new apps.

Patches (`patches:` with a strategic-merge or JSON 6902 patch) cover anything the shortcut fields
don't.
