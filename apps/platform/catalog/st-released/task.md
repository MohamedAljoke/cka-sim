---
id: st-released
title: Reuse a Released volume
domain: storage
host: cka-sim-control-plane
topics: [pv, claimref, reclaim-policy, pvc]
weight: 5
---
## Context

PV `pv-archive` (reclaim policy `Retain`) held the archive of an old app. Its claim was deleted, so
the PV is now `Released` with the data still on it. A new claim `archive-v2` in Namespace
`st-released`, used by Pod `reader`, stays `Pending`.

## Task

Make PVC `archive-v2` bind to PV `pv-archive`, so that Pod `reader` runs and can read the old
data in `/data/keep.txt`.

Do not delete or recreate PV `pv-archive`, and keep its reclaim policy.
