kubectl label namespace wl-psa pod-security.kubernetes.io/enforce=restricted pod-security.kubernetes.io/enforce-version=latest
kubectl -n wl-psa patch deployment web --type=strategic -p '
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: busybox
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: [ALL]
'
kubectl -n wl-psa rollout status deployment web --timeout=120s
