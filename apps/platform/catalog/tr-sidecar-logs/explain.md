`kubectl logs` only shows a container's **stdout and stderr**. An app that writes to a file inside
its container is invisible to it and to any cluster log collector. Without changing the app, add a
**sidecar** that shares the file through a volume and prints it:

```yaml
- name: log-streamer
  image: busybox:1.36
  command: [sh, -c, "tail -n+1 -F /var/log/orders/orders.log"]
  volumeMounts:
  - {name: logs, mountPath: /var/log/orders, readOnly: true}   # the same emptyDir as the app
```

```sh
kubectl -n tr-sidecar-logs edit deployment orders            # add the container
kubectl -n tr-sidecar-logs logs deploy/orders -c log-streamer -f
```

`tail -F` (capital F) follows the file by name, so it keeps working when the app rotates it. Both
containers mount the same `emptyDir`, which lives as long as the Pod. Since Kubernetes 1.29 the
sidecar can also be a native sidecar: an init container with `restartPolicy: Always`. It then
starts before the app and stops after it.
