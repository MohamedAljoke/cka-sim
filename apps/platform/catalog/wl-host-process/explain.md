A container isn't a VM. It's an ordinary Linux process on the node, isolated by **namespaces**
(its own PID, network, mount and UTS view) and limited by **cgroups**. From the host you see it
like any other process:

```sh
crictl ps --name proc
crictl inspect <id> | jq .info.pid               # host PID
ps -o pid,uid,user,args -p <pid>                 # uid 3000: the runAsUser you set
cat /proc/<pid>/status | grep -E 'Uid|NSpid'     # NSpid: 4321  1  <- PID 1 inside its own PID namespace
ps -ef --forest | grep -B2 'sleep 86401'         # its parent: the containerd-shim
```

That's why `runAsUser` matters so much (the course's "Security in Docker"): without user
namespaces, UID 0 in the container **is** root on the host, held back only by dropped capabilities,
seccomp and the rest. A container breakout as root is a node takeover; as UID 3000 it's an
unprivileged user. Pods can opt into user namespaces (`hostUsers: false`), so that root inside maps
to an unprivileged ID outside.
