kubectl -n wl-init patch deployment web --type=strategic -p '
spec:
  template:
    spec:
      initContainers:
      - name: wait-for-db
        image: busybox:1.36
        command: [sh, -c, "until nslookup db.wl-init.svc.cluster.local; do echo waiting for db; sleep 2; done"]
'
kubectl -n wl-init expose deployment db --port=80
kubectl -n wl-init rollout status deployment web --timeout=120s
