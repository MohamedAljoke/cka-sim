---
id: st-projected
title: One projected volume from four sources
domain: storage
host: cka-sim-control-plane
topics: [projected-volumes, configmaps, secrets, downward-api, serviceaccount-tokens]
weight: 5
---
## Context

Namespace `st-projected` has ConfigMap `app-config` (key `settings`) and Secret `app-secret` (key
`token`).

## Task

Create Pod `app` (image `busybox:1.36`, command `sleep 1d`, label `tier=backend`) with a single
`projected` volume mounted at `/etc/app` that contains:

| File | From |
|---|---|
| `/etc/app/settings` | key `settings` of ConfigMap `app-config` |
| `/etc/app/token` | key `token` of Secret `app-secret` |
| `/etc/app/labels` | the Pod's own labels (downward API) |
| `/etc/app/vault-token` | a ServiceAccount token for the audience `vault`, valid for `3600` seconds |
