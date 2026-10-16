An `emptyDir` belongs to one Pod. A Deployment never "restarts" a Pod: it replaces it with a new
one, with a new, empty `emptyDir`. Data that must outlive the Pod goes on a **PersistentVolumeClaim**,
which exists independently and is mounted by whichever Pod currently runs.

```yaml
volumes:
- name: data
  persistentVolumeClaim:
    claimName: notes-data      # instead of: emptyDir: {}
```

```sh
kubectl -n st-emptydir-to-pvc edit deployment notes
kubectl -n st-emptydir-to-pvc delete pod -l app=notes
kubectl -n st-emptydir-to-pvc exec deploy/notes -- cat /data/notes.txt   # lines from both Pods
```

With a `ReadWriteOnce` volume, use `strategy: Recreate` (as here) or make sure the old and new Pods
land on the same node. Otherwise a rolling update can deadlock: the new Pod waits for the volume the
old Pod still holds. For several replicas that each need their own data, use a StatefulSet with
`volumeClaimTemplates` instead of one shared PVC.
