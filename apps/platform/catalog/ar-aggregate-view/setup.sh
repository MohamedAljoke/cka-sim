kubectl delete clusterrole widget-viewer --ignore-not-found >/dev/null
kubectl delete crd widgets.cka.example.com --ignore-not-found --wait=true >/dev/null
fresh_ns ar-aggregate
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.cka.example.com
spec:
  group: cka.example.com
  scope: Namespaced
  names:
    kind: Widget
    plural: widgets
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        x-kubernetes-preserve-unknown-fields: true
YAML
kubectl wait --for=condition=Established crd/widgets.cka.example.com --timeout=60s >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: cka.example.com/v1
kind: Widget
metadata:
  name: gizmo
  namespace: ar-aggregate
spec:
  size: small
YAML
kubectl -n ar-aggregate create rolebinding reader-view --clusterrole=view --user=reader >/dev/null
