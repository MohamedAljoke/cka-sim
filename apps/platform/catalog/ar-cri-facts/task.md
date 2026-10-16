---
id: ar-cri-facts
title: Container runtime and crictl
domain: architecture
host: cka-sim-control-plane
topics: [cri, containerd, crictl]
weight: 3
---
## Task

On node `cka-sim-control-plane`, find out about the container runtime and write these lines to
`/opt/course/ar-cri-facts/runtime.txt`:

```
runtime: <name of the container runtime>
socket: <the runtime's CRI socket the kubelet uses>
kube-proxy: <ID of the running kube-proxy container on this node>
```

The container ID may be the short (13 characters) or the full form.
