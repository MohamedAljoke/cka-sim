Every Pod gets a **QoS class** in `status.qosClass`, computed from its containers' requests and
limits:

| Class | Rule | Evicted |
|---|---|---|
| `Guaranteed` | every container has CPU **and** memory limits, with requests equal to limits | last |
| `Burstable` | at least one request or limit set, but not Guaranteed | in between |
| `BestEffort` | no requests or limits at all, in any container | first |

```sh
kubectl get pods -A -o custom-columns=NS:.metadata.namespace,NAME:.metadata.name,QOS:.status.qosClass
kubectl get pods -A -o jsonpath='{range .items[?(@.status.qosClass=="BestEffort")]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}'
```

Read the class from the status rather than working it out from the spec: `metrics` here has only
a CPU **limit**, and since a limit without a request copies itself into the request, it's
Burstable. A LimitRange with defaults can also turn would-be BestEffort Pods into Burstable ones.
