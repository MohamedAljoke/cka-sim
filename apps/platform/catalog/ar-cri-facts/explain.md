The kubelet starts containers through a **container runtime** that speaks the CRI (Container
Runtime Interface): containerd or CRI-O on today's clusters. Docker Engine needs the extra
`cri-dockerd` shim since dockershim was removed in v1.24.

```sh
kubectl get nodes -o wide                                   # CONTAINER-RUNTIME column
grep containerRuntimeEndpoint /var/lib/kubelet/config.yaml  # the socket the kubelet uses
cat /etc/crictl.yaml                                        # the socket crictl uses
crictl ps                                                   # running containers on this node
crictl ps --name kube-proxy -q                              # just the ID
crictl pods                                                 # Pod sandboxes
crictl inspect <id> | jq .info.pid                          # the container's PID on the host
```

`crictl` is the node-level debugging tool. It shows what the runtime actually runs, including
static Pods, and still works when the API server is down. `crictl ps` hides exited containers;
add `-a` to see them. The short ID it prints (13 characters) works anywhere a full ID does.

`ctr` and `nerdctl` talk to containerd directly too, but they need `-n k8s.io` to see Kubernetes'
containers.
