kubectl apply -f /opt/course/ar-crd/crd.yaml
kubectl wait --for=condition=Established crd/backups.cka.example.com --timeout=30s
kubectl apply -f - <<'YAML'
apiVersion: cka.example.com/v1
kind: Backup
metadata:
  name: nightly
  namespace: ar-crd
spec:
  schedule: "0 2 * * *"
  retention: 7
YAML
kubectl explain backup.spec | awk '/^FIELDS:/{f=1;next} f && /^  [a-z]/{print $1}' > /opt/course/ar-crd/fields.txt
