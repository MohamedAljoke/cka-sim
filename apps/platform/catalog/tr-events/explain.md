Events are short-lived API objects (kept for 1 hour by default) that components write about
other objects: the scheduler writes `Scheduled`, the kubelet writes `Pulled`, `Started`,
`FailedMount` and `BackOff`. `kubectl describe pod` shows them at the bottom, but to save them
you list them directly.

`kubectl events` sorts by time and can filter on one object:

```sh
kubectl -n tr-events events --for pod/report > /opt/course/tr-events/events.txt
```

The older way works on every version and is worth knowing too:

```sh
kubectl -n tr-events get events --field-selector involvedObject.name=report --sort-by=.lastTimestamp
```

The trap is `kubectl get events` alone: it is not sorted by time, and it mixes in the events of
`ledger`. Repeated events are merged into one with a count (`x5 over 2m`), so a crash loop
doesn't flood the list.

Here the events show why `report` never starts: `MountVolume.SetUp failed ... configmap
"report-config" not found`. Creating that ConfigMap would let it start.
