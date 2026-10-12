kubectl delete crd backups.cka.example.com --ignore-not-found --wait=true >/dev/null
fresh_ns ar-crd
fresh_course
cat > "$COURSE/crd.yaml" <<'YAML'
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: backups.cka.example.com
spec:
  group: cka.example.com
  scope: Namespaced
  names:
    kind: Backup
    plural: backups
    singular: backup
    shortNames: [bk]
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            required: [schedule]
            properties:
              schedule:
                type: string
                description: Cron schedule of the backup.
              retention:
                type: integer
                description: How many backups to keep.
              target:
                type: string
                description: Where backups are stored.
YAML
chmod -R a+rwX "$COURSE"
