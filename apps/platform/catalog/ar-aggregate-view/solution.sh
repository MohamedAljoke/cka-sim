kubectl apply -f - <<'YAML'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: widget-viewer
  labels:
    rbac.authorization.k8s.io/aggregate-to-view: "true"
rules:
- apiGroups: [cka.example.com]
  resources: [widgets]
  verbs: [get, list, watch]
YAML
sleep 3
kubectl auth can-i list widgets.cka.example.com -n ar-aggregate --as=reader
