Minimal ("distroless") images ship only the app: no shell, no `ps`, no `ls`. `kubectl exec` has
nothing to run. An **ephemeral container** solves this. It's added to the running Pod (Pods can't
normally gain containers), uses whatever image you choose, and with `--target` it joins the target
container's **process namespace**, so it sees that container's processes.

```sh
kubectl -n tr-debug debug -it app --image=busybox:1.36 --container=debugger --target=app
#   / # ps          -> PID 1 /pause is the app container's process
#   / # ls /proc/1/root/      -> the app container's filesystem
```

Ephemeral containers can't be removed or restarted. They stay in `spec.ephemeralContainers` until
the Pod is deleted. They have no ports, probes or resources. Related tools:

- `kubectl debug app --copy-to=app-debug --share-processes ...` makes a debuggable **copy** of the
  Pod, for when you may change it (a different image or command).
- `kubectl debug node/<node> -it --image=busybox:1.36` runs a Pod on a node with the host's
  filesystem at `/host`, which helps when you have no ssh to the node.
