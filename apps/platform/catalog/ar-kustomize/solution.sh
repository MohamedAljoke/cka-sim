cat > /opt/course/ar-kustomize/overlays/prod/kustomization.yaml <<'YAML'
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../../base
namespace: ar-kustomize
replicas:
- name: shop
  count: 3
images:
- name: nginx
  newTag: 1.27-alpine
labels:
- pairs:
    env: prod
  includeTemplates: true
YAML
kubectl apply -k /opt/course/ar-kustomize/overlays/prod
kubectl -n ar-kustomize rollout status deployment shop --timeout=120s
