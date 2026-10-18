kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: meta
  namespace: st-downward-api
  labels:
    app: meta
  annotations:
    owner: team-a
spec:
  containers:
  - name: meta
    image: busybox:1.36
    command: [sleep, 1d]
    resources:
      limits:
        memory: 64Mi
    volumeMounts:
    - name: podinfo
      mountPath: /etc/podinfo
  volumes:
  - name: podinfo
    downwardAPI:
      items:
      - path: labels
        fieldRef: {fieldPath: metadata.labels}
      - path: annotations
        fieldRef: {fieldPath: metadata.annotations}
      - path: mem_limit
        resourceFieldRef:
          containerName: meta
          resource: limits.memory
          divisor: 1Mi
YAML
kubectl -n st-downward-api wait --for=condition=Ready pod/meta --timeout=60s
