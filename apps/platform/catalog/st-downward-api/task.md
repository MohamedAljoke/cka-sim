---
id: st-downward-api
title: Pod facts as files
domain: storage
host: cka-sim-control-plane
topics: [downward-api, volumes, resources]
weight: 3
---
## Task

In Namespace `st-downward-api`, create Pod `meta` (image `busybox:1.36`, command `sleep 1d`, label
`app=meta`, annotation `owner=team-a`, memory limit `64Mi`) with a `downwardAPI` volume mounted
at `/etc/podinfo` that provides:

| File | Content |
|---|---|
| `/etc/podinfo/labels` | the Pod's labels |
| `/etc/podinfo/annotations` | the Pod's annotations |
| `/etc/podinfo/mem_limit` | the container's memory limit, in **Mi** (so the file contains `64`) |
