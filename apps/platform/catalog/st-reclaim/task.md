---
id: st-reclaim
title: Keep a volume's data after its claim
domain: storage
host: cka-sim-control-plane
topics: [pv, reclaim-policy, pvc]
weight: 3
---
## Context

Pod `orders-db` in Namespace `st-reclaim` stores its data on PVC `orders`. The team is about to
move the app and must not lose the data if the PVC is deleted.

## Task

1. Find the PV bound to PVC `orders` and write these lines to `/opt/course/st-reclaim/pv.txt`:

   ```
   pv: <name of the PV>
   policy: <its current reclaim policy>
   ```

2. Change that PV so its data is kept when PVC `orders` is deleted.
