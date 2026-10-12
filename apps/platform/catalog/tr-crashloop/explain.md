`CrashLoopBackOff` means the container starts, exits, and the kubelet restarts it with a growing
delay (10s, 20s, 40s … up to 5 min). The reason is almost always in what the container printed
just before it died. `kubectl logs` shows the current container; `--previous` (`-p`) shows the
one that crashed, which is what you want once a new attempt has started:

```sh
kubectl -n tr-crashloop logs deploy/api --previous
kubectl -n tr-crashloop describe pod -l app=api        # Last State: Terminated, Exit Code: 1
kubectl -n tr-crashloop get configmap api-config -o yaml
```

The app wants `/etc/api/config.yaml`. A ConfigMap volume turns each key into a file, so the key
`config.yml` becomes `/etc/api/config.yml`: one letter off. The fix is the ConfigMap key, since
the Deployment must stay as it is:

```sh
kubectl -n tr-crashloop logs deploy/api --previous | grep FATAL > /opt/course/tr-crashloop/error.txt
kubectl -n tr-crashloop edit configmap api-config      # rename config.yml to config.yaml
kubectl -n tr-crashloop rollout restart deployment api
```

The kubelet updates ConfigMap volumes in running Pods by itself, but only after up to a minute
plus the current back-off. `rollout restart` gets a fresh Pod right away and doesn't change the
containers or volumes, only a restart annotation on the Pod template.

The trap is changing the Deployment instead (an `items:` mapping or a new path). That works, but
the task said not to.
