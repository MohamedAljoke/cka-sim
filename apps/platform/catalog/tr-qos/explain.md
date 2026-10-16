Under memory pressure the kubelet evicts Pods in QoS order: **BestEffort** first, then
**Burstable** Pods using more than they requested, and **Guaranteed** last. A Pod is Guaranteed
only when **every** container sets CPU **and** memory limits, with requests equal to the limits:

```yaml
resources:
  requests: {cpu: 200m, memory: 128Mi}
  limits:   {cpu: 200m, memory: 128Mi}      # setting only limits works too: requests copy them
```

```sh
kubectl -n tr-qos get pods -o custom-columns=NAME:.metadata.name,QOS:.status.qosClass
kubectl replace --force -f /opt/course/tr-qos/payments.yaml
kubectl -n tr-qos get pod payments -o jsonpath='{.status.qosClass}'
```

A Pod's QoS class is fixed when it's created. Resizing a running Pod in place can change its
values, but never its class, so changing the class means recreating the Pod. In a Deployment that
happens through a rollout.

Guaranteed Pods also qualify for exclusive CPU cores with the kubelet's static CPU manager policy,
which latency-sensitive apps use.
