---
id: tr-events
title: Events of a failing Pod
domain: troubleshooting
host: cka-sim-control-plane
topics: [events, kubectl, pods]
weight: 3
---
## Context

Pod `report` in Namespace `tr-events` never starts. Another Pod in that Namespace has problems of
its own.

## Task

Write the events of Pod `report`, and of nothing else, sorted by time with the oldest first, to
`/opt/course/tr-events/events.txt`.
