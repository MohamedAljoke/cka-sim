kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: st-projected
  labels:
    tier: backend
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
    volumeMounts:
    - name: all-in-one
      mountPath: /etc/app
  volumes:
  - name: all-in-one
    projected:
      sources:
      - configMap:
          name: app-config
          items: [{key: settings, path: settings}]
      - secret:
          name: app-secret
          items: [{key: token, path: token}]
      - downwardAPI:
          items:
          - path: labels
            fieldRef: {fieldPath: metadata.labels}
      - serviceAccountToken:
          audience: vault
          expirationSeconds: 3600
          path: vault-token
YAML
kubectl -n st-projected wait --for=condition=Ready pod/app --timeout=60s
