for crd in dashboards.monitoring.cka.io alertrules.monitoring.cka.io probes.monitoring.cka.io panels.grafana.cka.io; do
  kubectl delete crd "$crd" --ignore-not-found --wait=true >/dev/null
done
fresh_ns ar-crd-list-team-a
fresh_ns ar-crd-list-team-b
fresh_course
crd() {
  cat <<YAML
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: $2.$1
spec:
  group: $1
  scope: Namespaced
  names:
    kind: $3
    plural: $2
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        x-kubernetes-preserve-unknown-fields: true
---
YAML
}
{
  crd monitoring.cka.io dashboards Dashboard
  crd monitoring.cka.io alertrules AlertRule
  crd monitoring.cka.io probes Probe
  crd grafana.cka.io panels Panel
} | kubectl apply -f - >/dev/null
kubectl wait --for=condition=Established crd/dashboards.monitoring.cka.io crd/panels.grafana.cka.io --timeout=60s >/dev/null
cr() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: $1/v1
kind: $2
metadata:
  name: $4
  namespace: $3
spec: {}
YAML
}
cr monitoring.cka.io Dashboard ar-crd-list-team-a cluster-overview
cr monitoring.cka.io Dashboard ar-crd-list-team-a node-health
cr monitoring.cka.io Dashboard ar-crd-list-team-b payments
cr monitoring.cka.io AlertRule ar-crd-list-team-b high-latency
cr grafana.cka.io Panel ar-crd-list-team-b dashboard-cpu
