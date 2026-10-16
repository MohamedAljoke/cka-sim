---
id: ar-etcd-facts
title: Facts about etcd
domain: architecture
host: cka-sim-control-plane
topics: [etcd, etcdctl, static-pods]
weight: 3
---
## Task

Find out the following about this cluster's etcd and write them to
`/opt/course/ar-etcd-facts/etcd.txt`:

```
version: <etcd server version, e.g. 3.5.12>
data-dir: <etcd's data directory on the node>
members: <names of all etcd members, comma-separated>
```
