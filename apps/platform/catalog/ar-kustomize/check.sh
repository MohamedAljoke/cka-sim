NS=ar-kustomize
D() { kubectl -n $NS get deployment shop -o jsonpath="$1"; }
deployed() { eq 3 "$(D '{.status.readyReplicas}')" && eq nginx:1.27-alpine "$(D '{.spec.template.spec.containers[0].image}')"; }
labelled() { eq prod "$(D '{.metadata.labels.env}')" && eq prod "$(D '{.spec.template.metadata.labels.env}')"; }
rendered() { kubectl kustomize "$COURSE/overlays/prod" | grep -q 'namespace: ar-kustomize' && kubectl kustomize "$COURSE/overlays/prod" | grep -q 'replicas: 3'; }
base_kept() { grep -q 'image: nginx:1.26-alpine' "$COURSE/base/deployment.yaml" && grep -q 'replicas: 1' "$COURSE/base/deployment.yaml"; }
untouched_base() { rendered && base_kept; }
deployed || wait_for 60 deployed

check 2 "Deployment shop runs 3 Ready replicas of nginx:1.27-alpine in ar-kustomize" deployed
check 1 "The Deployment and its Pods carry env=prod" labelled
check 2 "overlays/prod renders the namespace and 3 replicas" rendered
check 1 "The base files are unchanged" untouched_base
