kubectl -n wl-graceful patch deployment web --type=strategic -p '
spec:
  template:
    spec:
      terminationGracePeriodSeconds: 45
      containers:
      - name: nginx
        lifecycle:
          preStop:
            exec:
              command: [sleep, "10"]
'
kubectl -n wl-graceful rollout status deployment web --timeout=120s
