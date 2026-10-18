kubectl apply -f - <<'YAML'
apiVersion: v1
kind: LimitRange
metadata:
  name: defaults
  namespace: wl-quota
spec:
  limits:
  - type: Container
    defaultRequest:
      cpu: 100m
      memory: 64Mi
    default:
      cpu: 200m
      memory: 128Mi
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-quota
  namespace: wl-quota
spec:
  hard:
    pods: "4"
    requests.cpu: "1"
    limits.memory: 1Gi
YAML
kubectl -n wl-quota run greedy --image=busybox:1.36 --overrides='{"spec":{"containers":[{"name":"greedy","image":"busybox:1.36","command":["sleep","1d"],"resources":{"requests":{"cpu":"2"},"limits":{"cpu":"2"}}}]}}' 2> /opt/course/wl-quota/rejected.txt || true
