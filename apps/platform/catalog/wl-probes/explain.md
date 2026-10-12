Three probes, three questions the kubelet keeps asking:

| Probe | Question | On failure |
|---|---|---|
| **startup** | Has the app finished starting? | Keep waiting, up to `failureThreshold × periodSeconds`, then restart. Holds off the other two until it passes. |
| **liveness** | Is it still alive? | Restart the container. |
| **readiness** | Can it take traffic right now? | Remove the Pod from Service endpoints. No restart. |

Each probe uses one of three mechanisms: `httpGet` (any 2xx or 3xx is a pass), `tcpSocket` (the
port accepts a connection), or `exec` (the command exits 0). There is also `grpc`.

```yaml
startupProbe:
  httpGet: {path: /, port: 80}
  periodSeconds: 2
  failureThreshold: 30        # up to 60 s to start
livenessProbe:
  tcpSocket: {port: 80}
  periodSeconds: 10
readinessProbe:
  exec:
    command: [cat, /usr/share/nginx/html/index.html]
  periodSeconds: 5
```

The startup probe exists for slow starters: without it, a liveness probe tuned for a running app
would kill the container before it ever finished booting. A common trap is a liveness probe that
checks a dependency such as the database. When the database blips, every Pod restarts at once. Put
dependency checks in readiness, if anywhere.
