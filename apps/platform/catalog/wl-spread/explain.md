`topologySpreadConstraints` tell the scheduler to keep matching Pods balanced across
**topology domains**, one per distinct value of `topologyKey`. With `kubernetes.io/hostname`,
each node is a domain. `maxSkew: 1` means the busiest and the emptiest domain may differ by at
most one Pod. `whenUnsatisfiable: DoNotSchedule` makes it a hard rule (`ScheduleAnyway` only
prefers it).

```yaml
topologySpreadConstraints:
- maxSkew: 1
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: DoNotSchedule
  nodeTaintsPolicy: Honor
  labelSelector:
    matchLabels:
      app: web
  matchLabelKeys: [pod-template-hash]
```

The trap is the **tainted control plane**. By default the constraint counts every node as a
domain, including ones the Pod can't run on because of a taint. The control plane then counts as
a domain with 0 Pods, so the workers may only get 1 Pod each, and 2 replicas stay `Pending` with
`didn't match pod topology spread constraints`. `nodeTaintsPolicy: Honor` leaves out nodes whose
taints the Pod doesn't tolerate. A `nodeAffinity` limited to the workers works too.

Then the **rollout** trap: while the change rolls out, the old Pods (all on one node) still match
`app: web` and are counted, so the new Pods get balanced against Pods that are about to disappear.
You end up 3:1. `matchLabelKeys: [pod-template-hash]` adds each Pod's own template hash to the
selector, so only Pods of the same revision are counted.

Don't forget the `labelSelector`. It decides which Pods are counted, and without it the
constraint counts nothing. Also remove the old preferred affinity that pulled every Pod to one
node. Spread is only checked when a Pod is scheduled, so the change rolls out new Pods.
