kubectl apply -f - <<'YAML'
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: support
  namespace: ar-subresources
rules:
- apiGroups: [""]
  resources: [pods, pods/log]
  verbs: [get, list]
- apiGroups: [""]
  resources: [pods/exec]
  verbs: [create]
- apiGroups: [apps]
  resources: [deployments/scale]
  verbs: [get, update, patch]
YAML
kubectl -n ar-subresources create rolebinding support --role=support --group=support
