---
id: wl-configmap
title: ConfigMap as a file and an env var
domain: workloads
host: cka-sim-control-plane
topics: [configmaps, volumes, env]
weight: 4
---
## Task

In Namespace `wl-configmap`:

1. Create ConfigMap `app-config` with the key `LOG_LEVEL=debug` (from a literal) and the key
   `app.properties` with the content of the file `/opt/course/wl-configmap/app.properties`.
2. Create Pod `app` (image `busybox:1.36`, command `sleep 1d`) that:
   - mounts only the key `app.properties` as the file `/etc/app/app.properties`
   - gets the env var `LOG_LEVEL` from the ConfigMap key `LOG_LEVEL`
