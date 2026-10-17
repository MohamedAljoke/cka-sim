default() { kubectl get storageclass "$1" -o jsonpath='{.metadata.annotations.storageclass\.kubernetes\.io/is-default-class}' 2>/dev/null; }
only_standard() { eq true "$(default standard)" && kubectl get storageclass legacy >/dev/null 2>&1 && [ "$(default legacy)" != true ]; }
answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/classes.txt" 2>/dev/null | head -1 | tr -d ' \r"'; }
gets() {
  kubectl create --dry-run=server -o jsonpath='{.spec.storageClassName}' -f - 2>/dev/null <<YAML
apiVersion: v1
kind: PersistentVolumeClaim
metadata: {name: probe, namespace: default}
spec:
  accessModes: [ReadWriteOnce]
  resources: {requests: {storage: 1Mi}}
  $1
YAML
}
new_pvcs() { only_standard && eq standard "$(gets '')"; }

check 2 "standard is the only default; legacy still exists" only_standard
check 1 "A new PVC without a class now gets standard" new_pvcs
check 1 "classes.txt: omitted gets standard" eq standard "$(answer omitted)"
check 1 "classes.txt: an empty storageClassName gets none" eq none "$(answer empty)"
