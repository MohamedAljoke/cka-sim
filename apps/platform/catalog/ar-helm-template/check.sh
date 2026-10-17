rendered() { grep -q '^kind: Deployment' "$COURSE/docs.yaml" && grep -q 'name: docs-web' "$COURSE/docs.yaml" && grep -q 'image: "nginx:1.26-alpine"' "$COURSE/docs.yaml"; }
created() { eq "1 nginx:1.26-alpine" "$(kubectl -n ar-helm-template get deployment docs-web -o jsonpath='{.status.readyReplicas} {.spec.template.spec.containers[0].image}' 2>/dev/null)"; }
service() { kubectl -n ar-helm-template get service docs-web >/dev/null 2>&1; }
no_release() { [ -z "$(helm -n ar-helm-template ls -a -q)" ] && [ -z "$(kubectl -n ar-helm-template get secrets -l owner=helm -o name)" ]; }

check 2 "docs.yaml is chart web 1.0.0 rendered as docs" rendered
check 4 "Deployment docs-web and Service docs-web exist, with no Helm release in ar-helm-template" eval 'wait_for 60 created && service && no_release'
