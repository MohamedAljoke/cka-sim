`kubectl logs` can read many Pods at once through a label selector, and `--prefix` tags every line
with its Pod and container:

```sh
kubectl -n tr-logs-count logs -l app=worker --prefix | grep ERROR
kubectl -n tr-logs-count logs deploy/worker                 # only ONE Pod of the Deployment
kubectl -n tr-logs-count logs -l app=worker --prefix --since=1h --tail=-1
```

Two traps: `logs deploy/worker` quietly picks a single Pod, and `-l` limits each Pod to its last 10
lines unless you add `--tail=-1`. For an exact count per Pod, a loop is clearest:

```sh
for p in $(kubectl -n tr-logs-count get pods -l app=worker -o name); do
  echo "${p#pod/}: $(kubectl -n tr-logs-count logs $p | grep -c ERROR)"
done > /opt/course/tr-logs-count/errors.txt
```

`grep -c` counts matching lines and prints `0` when nothing matches, so every Pod gets a line.
