**Init containers** run before the app containers, one after another, and each must exit 0
before the next starts. If one fails, the kubelet retries it. The Pod shows `Init:0/1` until they
are all done. That makes them the natural place for "wait until X is ready" and for one-time
setup work (migrations, fetching config, fixing volume permissions).

```yaml
initContainers:
- name: wait-for-db
  image: busybox:1.36
  command: [sh, -c, "until nslookup db.wl-init.svc.cluster.local; do echo waiting; sleep 2; done"]
```

```sh
kubectl -n wl-init get pods                         # Init:0/1 while db has no Service
kubectl -n wl-init logs <web-pod> -c wait-for-db    # "waiting"
kubectl -n wl-init expose deployment db --port=80
kubectl -n wl-init get pods -w                      # PodInitializing, then Running
```

The DNS name only resolves once the Service exists, which makes it a good readiness signal for
"someone has published this". Init containers don't support liveness or readiness probes; they
just have to finish. For helpers that must keep running beside the app, use a sidecar instead
(an init container with `restartPolicy: Always`).
