---
id: ar-kubelet-facts
title: Where the kubelet keeps its config
domain: architecture
host: cka-sim-worker2
topics: [kubelet, systemd, static-pods, cgroups]
weight: 3
---
## Task

On node `cka-sim-worker2`, find out how the kubelet is configured and write these lines to
`/opt/course/ar-kubelet-facts/kubelet.txt`:

```
config: <path of the kubelet's config file>
staticPodPath: <directory the kubelet reads static Pod manifests from>
cgroupDriver: <cgroup driver the kubelet uses>
```
