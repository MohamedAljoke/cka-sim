kubectl delete crd gadgets.cka.example.com --ignore-not-found --wait=true >/dev/null
fresh_ns ar-crd-versions
fresh_course
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: gadgets.cka.example.com
spec:
  group: cka.example.com
  scope: Namespaced
  names:
    kind: Gadget
    plural: gadgets
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              color:
                type: string
YAML
kubectl wait --for=condition=Established crd/gadgets.cka.example.com --timeout=60s >/dev/null
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: cka.example.com/v1alpha1
kind: Gadget
metadata:
  name: prototype
  namespace: ar-crd-versions
spec:
  color: red
YAML
