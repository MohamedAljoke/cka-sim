The `openAPIV3Schema` of a CRD version is enforced by the API server on every create and update.
Bad objects never reach etcd:

```yaml
schema:
  openAPIV3Schema:
    type: object
    properties:
      spec:
        type: object
        required: [engine]
        properties:
          engine:    {type: string, enum: [postgres, mysql]}
          storageGB: {type: integer, minimum: 1}
```

```sh
kubectl -n ar-crd-schema apply -f legacy.yaml
#   The Database "legacy" is invalid: spec.engine: Unsupported value: "oracle": supported values: "postgres", "mysql"
kubectl -n ar-crd-schema get db                      # NAME  ENGINE  STORAGE  AGE
kubectl explain database.spec.engine
```

`additionalPrinterColumns` choose what `kubectl get` shows, so users see the important fields
without `-o yaml`. `shortNames` give kubectl aliases (`kubectl get db`).

Validation rules beyond types, enums and ranges can be written in CEL with
`x-kubernetes-validations` (for example "storageGB may only grow"). Fields not in the schema are
**pruned** silently unless you mark a subtree `x-kubernetes-preserve-unknown-fields: true`.
