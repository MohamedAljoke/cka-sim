---
id: tr-static-yaml
title: Static Pod never appears
domain: troubleshooting
host: cka-sim-worker2
topics: [static-pods, kubelet, journalctl, yaml]
weight: 4
---
## Context

An admin placed a manifest for the static Pod `node-exporter` on node `cka-sim-worker2`, but the Pod
never shows up, not even in `crictl ps -a`.

## Task

Find out why and fix the manifest, so that static Pod `node-exporter` runs on `cka-sim-worker2`
(its mirror Pod appears in Namespace `tr-static-yaml`). Keep its image and command.
