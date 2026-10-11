`kubectl scale` changes `.spec.replicas` on the Deployment. The Deployment controller passes the
new count to its ReplicaSet, and the ReplicaSet controller creates the missing Pods.

```sh
kubectl -n wl-scale scale deployment web --replicas=4
kubectl -n wl-scale rollout status deployment/web
```

`kubectl edit deployment web` and changing `replicas:` works too, but `scale` is faster in the exam.
Setting replicas only asks for 4 Pods; `rollout status` (or `get deploy`) shows when all 4 are Ready.
