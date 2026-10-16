Two admission controllers act on every Pod a namespace receives:

- **LimitRange** *changes* Pods: it fills in default requests and limits for containers that set
  none, and it can refuse values outside a `min`/`max`.
- **ResourceQuota** *counts* Pods: it adds up the requests, limits and objects in the namespace
  and rejects anything that would go over `hard`.

```sh
kubectl -n wl-quota create quota team-quota --hard=pods=4,requests.cpu=1,limits.memory=1Gi
kubectl -n wl-quota describe quota team-quota          # Used vs Hard
kubectl -n wl-quota describe limitrange defaults
kubectl -n wl-quota run greedy --image=busybox:1.36 --overrides='...requests cpu 2...' 2> /opt/course/wl-quota/rejected.txt
#   pods "greedy" is forbidden: exceeded quota: team-quota, requested: requests.cpu=2, used: ..., limited: requests.cpu=1
```

The two work together. Once a quota covers `requests.cpu` or `limits.memory`, every new Pod
**must** set those values, or it's rejected too. A LimitRange with defaults saves every team from
writing them by hand. There is no imperative command for a LimitRange, so take the YAML from the
docs page "Limit Ranges".

Watch the order: the LimitRange runs first. A Pod that requests `cpu: 2` but sets no limit gets
the default limit of `200m`, and the request is then larger than the limit. That Pod is refused
for being invalid, not by the quota. That's why `greedy` sets its limit too.

A quota only checks new Pods. Pods that were already running stay, even if they now exceed it.
For Deployments, look at the ReplicaSet's events: the Deployment looks fine, but its Pods are
refused.
