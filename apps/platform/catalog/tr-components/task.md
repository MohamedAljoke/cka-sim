---
id: tr-components
title: How the components run
domain: troubleshooting
host: cka-sim-control-plane
topics: [static-pods, daemonsets, systemd, control-plane]
weight: 3
---
## Task

Find out how each of these components is run in this cluster, and write one line per component
to `/opt/course/tr-components/components.txt`:

```
kubelet: <how>
kube-apiserver: <how>
kube-scheduler: <how>
kube-controller-manager: <how>
etcd: <how>
coredns: <how>
kube-proxy: <how>
```

`<how>` is one of `systemd`, `static-pod`, `deployment` or `daemonset`.
