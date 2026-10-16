A Deployment is only a wish. The **Deployment controller** creates its ReplicaSet, and the
**ReplicaSet controller** creates the Pods. Both run inside **kube-controller-manager**, together
with the node, endpoint, namespace, job and garbage-collector controllers. While it is down the
API server happily stores new objects, but nothing acts on them: no ReplicaSet, no Pods, no
scaling.

```sh
kubectl -n tr-controller-manager get deploy,rs,pods
kubectl -n kube-system get pods | grep controller-manager       # CrashLoopBackOff
kubectl -n kube-system logs kube-controller-manager-cka-sim-control-plane
grep kubeconfig /etc/kubernetes/manifests/kube-controller-manager.yaml
ls /etc/kubernetes/*.conf                                       # controller-manager.conf exists
vim /etc/kubernetes/manifests/kube-controller-manager.yaml
```

kube-controller-manager is a static Pod like the scheduler and the API server, so you fix the file
in `/etc/kubernetes/manifests/`, not the mirror Pod in `kube-system`. Once it is back it picks up
the existing Deployment and creates the ReplicaSet. There is no need to touch `web`. Recreating
`web` would hide the evidence and break the rule.

Tell the symptoms apart: Pods `Pending` with no node means the **scheduler**; no ReplicaSet or
Pods at all means the **controller-manager**.
