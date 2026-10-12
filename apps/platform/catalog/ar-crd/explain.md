A **CustomResourceDefinition** teaches the API server a new kind. Once it's `Established`, the
kind gets a REST path (`/apis/cka.example.com/v1/namespaces/*/backups`) and kubectl works with
it like a built-in: `get`, `describe`, `-o yaml`, `explain`, and its short name `bk`.

```sh
kubectl apply -f /opt/course/ar-crd/crd.yaml
kubectl get crd backups.cka.example.com
kubectl api-resources --api-group=cka.example.com
kubectl explain backup.spec                # fields come from the CRD's openAPIV3Schema
kubectl -n ar-crd get backups              # or: kubectl -n ar-crd get bk
```

The `openAPIV3Schema` validates every custom resource: `retention: "7"` (a string) or a missing
`schedule` is rejected by the API server. `kubectl explain` reads the same schema, which makes it
the quickest way to learn an unfamiliar kind in the exam.

A CRD alone only stores data. Nothing acts on a Backup until a **controller** (often called an
operator) watches Backups and does the work. Deleting the CRD deletes every custom resource of
that kind with it.
