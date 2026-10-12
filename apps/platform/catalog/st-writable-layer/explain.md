A container's filesystem is the image's read-only layers plus a thin **writable layer** on top,
just like `docker run`. That layer belongs to the container instance: when the container
restarts (a crash, a failed liveness probe, `kill 1`), the runtime creates a new container from
the image, and everything written before is gone.

Volumes live outside that layer. An `emptyDir` belongs to the **Pod**: it survives container
restarts and is deleted with the Pod. For data that must outlive the Pod, use a PersistentVolume.

| Where data is written | Survives container restart | Survives Pod deletion |
|---|---|---|
| container writable layer | no | no |
| `emptyDir` | yes | no |
| PVC / PersistentVolume | yes | yes |

```sh
kubectl -n st-writable-layer exec scratch -- cat /cache/marker
kubectl -n st-writable-layer exec scratch -- kill 1        # PID 1 exits, the kubelet restarts the container
kubectl -n st-writable-layer get pod scratch               # RESTARTS 1
kubectl -n st-writable-layer exec scratch -- cat /cache/marker      # same timestamp with emptyDir
```

A volume can't be added to a running Pod, which is why the Pod is recreated from the file.
