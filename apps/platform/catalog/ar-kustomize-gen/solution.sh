cd /opt/course/ar-kustomize-gen
cat > kustomization.yaml <<'YAML'
namespace: ar-kustomize-gen
resources:
- deployment.yaml
configMapGenerator:
- name: app-config
  files:
  - app.properties
secretGenerator:
- name: app-secret
  literals:
  - password=s3cr3t
YAML
kubectl apply -k .
kubectl -n ar-kustomize-gen rollout status deployment app --timeout=120s
sed -i 's/^mode=prod$/mode=debug/' app.properties
kubectl apply -k .
kubectl -n ar-kustomize-gen rollout status deployment app --timeout=120s
