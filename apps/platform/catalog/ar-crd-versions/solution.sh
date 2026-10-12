kubectl get crd gadgets.cka.example.com -o json |
  jq '.spec.versions[0].storage = false | .spec.versions += [.spec.versions[0] | .name = "v1" | .storage = true]' |
  kubectl replace -f -
kubectl apply -f - <<'YAML'
apiVersion: cka.example.com/v1
kind: Gadget
metadata:
  name: beacon
  namespace: ar-crd-versions
spec:
  color: blue
YAML
kubectl get crd gadgets.cka.example.com -o jsonpath='{.status.storedVersions}' | jq -r 'join(",")' > /opt/course/ar-crd-versions/stored.txt
