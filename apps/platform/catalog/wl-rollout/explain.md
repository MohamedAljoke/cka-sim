**Concept.** A Deployment never edits Pods in place. Each change to its Pod template creates a
new **ReplicaSet**, and the rolling update scales the new one up while scaling the old one down.
Old ReplicaSets are kept at 0 replicas — that is the revision history `rollout undo` uses.

**Rollback creates a new revision.** `rollout undo` copies revision 2's template forward, so the
Deployment ends up on revision **4**, and revision 2 disappears from the history. The number is
in the annotation `deployment.kubernetes.io/revision` (and the last line of `rollout history`).

**Strategy knobs:** `maxUnavailable` = how many Pods may be missing during the update;
`maxSurge` = how many extra Pods may exist above `replicas`. Both accept numbers or percentages.
Fastest edit on the exam: `k edit deploy shop` and change `spec.strategy.rollingUpdate`.
