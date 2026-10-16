---
id: ar-kustomize-gen
title: Kustomize ConfigMap and Secret generators
domain: architecture
host: cka-sim-control-plane
topics: [kustomize, configmaps, secrets, rollouts]
weight: 5
---
## Context

`/opt/course/ar-kustomize-gen/` holds `deployment.yaml` for Deployment `app` and its settings in
`app.properties`. The Deployment reads ConfigMap `app-config` and Secret `app-secret`.

## Task

1. Write `/opt/course/ar-kustomize-gen/kustomization.yaml` that deploys `deployment.yaml` into
   Namespace `ar-kustomize-gen`, generates ConfigMap `app-config` from the file
   `app.properties`, and generates Secret `app-secret` with the literal `password=s3cr3t`.
2. Apply it with `kubectl apply -k`.
3. Then change `mode=prod` to `mode=debug` in `app.properties` and apply again, so that the
   Deployment rolls out new Pods that see the new file.
