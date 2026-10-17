sc() { kubectl get storageclass fast-local -o jsonpath="$1"; }
shape() { eq "rancher.io/local-path Retain WaitForFirstConsumer" "$(sc '{.provisioner} {.reclaimPolicy} {.volumeBindingMode}')"; }
default() { kubectl get storageclass "$1" -o jsonpath='{.metadata.annotations.storageclass\.kubernetes\.io/is-default-class}'; }
only_default() { eq true "$(default fast-local)" && [ "$(default standard)" != true ]; }
gets_it() {
  kubectl create -f - --dry-run=server -o jsonpath='{.spec.storageClassName}' 2>/dev/null <<'YAML' | grep -qx fast-local
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: probe
  namespace: default
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 1Mi
YAML
}

check 2 "fast-local uses rancher.io/local-path with Retain and WaitForFirstConsumer" shape
check 2 "fast-local is the only default StorageClass" only_default
check 1 "A new PVC without a class gets fast-local" gets_it
