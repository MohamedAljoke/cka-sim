Containers are ordinary processes on the node, and Pod volumes are ordinary directories there.
The kubelet keeps each Pod's volumes under its Pod UID:

```
/var/lib/kubelet/pods/<pod-uid>/volumes/<plugin>/<volume-name>
   kubernetes.io~empty-dir/scratch
   kubernetes.io~configmap/...
   kubernetes.io~projected/kube-api-access-xxxxx      # the ServiceAccount token
```

```sh
kubectl -n st-node-paths get pod cache -o jsonpath='{.metadata.uid}'
ls /var/lib/kubelet/pods/<uid>/volumes/kubernetes.io~empty-dir/scratch/
crictl ps --name cache                                    # the container ID
crictl inspect <id> | jq .info.pid                        # its main process on the host
ps -o pid,user,args -p <pid>
nsenter -t <pid> -m -- ls /scratch                        # the same files, from inside its mount namespace
```

This is what "Docker storage" looks like under Kubernetes: the image layers sit in containerd's
snapshotter (`/var/lib/containerd/`), and volumes are bind-mounted into the container from the
kubelet's directory. When the Pod is deleted, the kubelet removes its directory, emptyDir data
included.
