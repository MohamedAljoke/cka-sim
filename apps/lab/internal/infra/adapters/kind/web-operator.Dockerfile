FROM cka-sim/node:v1.37.0 AS node

FROM busybox:1.36
COPY --from=node /usr/bin/kubectl /usr/local/bin/kubectl
COPY <<'EOF' /operator.sh
seen=""
while true; do
  current=$(kubectl get webapps -A -o jsonpath='{range .items[*]}{.metadata.namespace} {.metadata.name} {.metadata.uid} {.spec.image} {.spec.replicas} {.spec.port}{"\n"}{end}')
  echo "$current" | while read -r ns name uid image replicas port; do
    [ -n "$name" ] || continue
    kubectl apply --server-side --force-conflicts --field-manager=web-operator -f - >/dev/null <<YAML || echo "failed to reconcile $ns/$name"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $name
  namespace: $ns
  labels: {app.kubernetes.io/managed-by: web-operator, webapp: $name}
  ownerReferences: [{apiVersion: cka-sim.io/v1, kind: WebApp, name: $name, uid: $uid, controller: true}]
spec:
  replicas: $replicas
  selector: {matchLabels: {webapp: $name}}
  template:
    metadata: {labels: {webapp: $name}}
    spec:
      containers: [{name: web, image: "$image", ports: [{containerPort: $port}]}]
---
apiVersion: v1
kind: Service
metadata:
  name: $name
  namespace: $ns
  labels: {app.kubernetes.io/managed-by: web-operator, webapp: $name}
  ownerReferences: [{apiVersion: cka-sim.io/v1, kind: WebApp, name: $name, uid: $uid, controller: true}]
spec:
  selector: {webapp: $name}
  ports: [{port: $port, targetPort: $port}]
YAML
    ready=$(kubectl -n "$ns" get deployment "$name" -o jsonpath='{.status.readyReplicas}')
    kubectl -n "$ns" patch webapp "$name" --subresource=status --type=merge -p "{\"status\":{\"readyReplicas\":${ready:-0}}}" >/dev/null
  done
  [ "$current" = "$seen" ] || echo "reconciled: $(echo "$current" | awk 'NF {printf "%s/%s ", $1, $2}')"
  seen=$current
  sleep 5
done
EOF
ENTRYPOINT ["sh", "/operator.sh"]
