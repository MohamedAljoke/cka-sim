This is how cluster logging works (Fluent Bit, Vector, Promtail, the OpenTelemetry collector): a
**DaemonSet** runs one agent per node, mounts the node's log directory with a **hostPath**
volume, and ships what it reads elsewhere. The container runtime writes every container's output
to:

```
/var/log/pods/<namespace>_<pod>_<pod-uid>/<container>/0.log
```

The directory name starts with the namespace, so a glob selects one namespace. Each line carries a
prefix added by the runtime: `<timestamp> stdout F <message>`.

```sh
ls /var/log/pods/ | grep tr-log-agent-apps                         # on the node
kubectl -n tr-log-agent get pods -o wide                           # one agent per worker
kubectl -n tr-log-agent logs <agent-on-app's-node> --tail=5
```

Mount it `readOnly: true`: an agent only needs to read, and a hostPath into node directories is a
powerful thing to hand to a Pod. `/var/log/containers/` holds symlinks with a flatter name, and
real agents usually read that.

The control plane is tainted, so a DaemonSet without a matching toleration runs only on the
workers, which is what this task asks for.
