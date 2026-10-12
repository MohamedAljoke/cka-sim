A Deployment has two strategies:

- **RollingUpdate** (the default) replaces Pods a few at a time, as `maxSurge` and
  `maxUnavailable` allow. Old and new versions run side by side for a while, with no downtime.
- **Recreate** scales the old ReplicaSet to 0, waits until its Pods are gone, then starts the new
  ones. There is downtime, but never two versions at once.

Choose Recreate when versions can't coexist: a schema migration, an app that holds an exclusive
lock or a `ReadWriteOnce` volume that only one Pod can mount, or a licence that allows one
instance.

```sh
kubectl -n wl-recreate edit deployment db-migrator
#   strategy:
#     type: Recreate          # and delete the rollingUpdate: block
kubectl -n wl-recreate set image deployment db-migrator nginx=busybox:1.36
kubectl -n wl-recreate get pods -w          # all old Pods Terminating, then new ones
```

The trap is setting `type: Recreate` while `rollingUpdate:` is still there. The API server
rejects it with `may not be specified when strategy type is 'Recreate'`. A merge patch must set
`rollingUpdate: null`.
