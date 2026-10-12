kubectl -n wl-probes patch deployment api --type=strategic -p '
spec:
  template:
    spec:
      containers:
      - name: nginx
        startupProbe:
          httpGet: {path: /, port: 80}
          periodSeconds: 2
          failureThreshold: 30
        livenessProbe:
          tcpSocket: {port: 80}
          periodSeconds: 10
        readinessProbe:
          exec:
            command: [cat, /usr/share/nginx/html/index.html]
          periodSeconds: 5
'
kubectl -n wl-probes rollout status deployment api --timeout=120s
