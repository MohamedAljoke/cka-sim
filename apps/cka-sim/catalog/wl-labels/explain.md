Labels are key/value pairs that selectors match. Services, Deployments, ReplicaSets and NetworkPolicies
all find their Pods by label, so changing a label can move a Pod in or out of a Service. Annotations
look the same but nothing selects on them: they hold notes for people and tools.

```sh
kubectl -n wl-labels get pods --show-labels
kubectl -n wl-labels get pods -l tier=backend,env=prod -o name | cut -d/ -f2 > /opt/course/wl-labels/backend-prod.txt
kubectl -n wl-labels label pods -l app=web team=payments
kubectl -n wl-labels label pod web-dev env-
kubectl -n wl-labels annotate pod api owner=platform
```

A comma in `-l` means AND. Set-based selectors cover the rest: `-l 'env in (dev,prod)'`,
`-l 'env notin (prod)'` (this also matches Pods with no `env` label), `-l '!env'` and `-l env`.
`-o name` prints `pod/api`, hence the `cut`; `-o custom-columns=:metadata.name --no-headers` works too.
The trailing `-` in `env-` removes a label; adding one that already exists needs `--overwrite`.
