kubectl apply -f - <<'YAML'
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.cka.example.com
spec:
  group: cka.example.com
  scope: Namespaced
  names:
    kind: Database
    plural: databases
    singular: database
    shortNames: [db]
  versions:
  - name: v1
    served: true
    storage: true
    additionalPrinterColumns:
    - name: Engine
      type: string
      jsonPath: .spec.engine
    - name: Storage
      type: integer
      jsonPath: .spec.storageGB
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            required: [engine]
            properties:
              engine:
                type: string
                enum: [postgres, mysql]
              storageGB:
                type: integer
                minimum: 1
YAML
kubectl wait --for=condition=Established crd/databases.cka.example.com --timeout=30s
db() { printf 'apiVersion: cka.example.com/v1\nkind: Database\nmetadata:\n  name: %s\n  namespace: ar-crd-schema\nspec:\n  engine: %s\n  storageGB: 10\n' "$1" "$2"; }
db orders postgres | kubectl apply -f -
db legacy oracle | kubectl apply -f - 2> /opt/course/ar-crd-schema/rejected.txt || true
kubectl -n ar-crd-schema get db
