`--sort-by` takes a JSONPath into each object and sorts the list by it, ascending. Anything in
`kubectl get pod <name> -o yaml` can be a key:

```sh
kubectl get pods -A --sort-by=.metadata.creationTimestamp
kubectl -n tr-sort-by get pods --sort-by='.status.containerStatuses[0].restartCount'
```

Creation timestamps sort oldest first, so the newest Pod ends up at the bottom. For restarts
the biggest number is last; `tail -1` picks it. Don't trust the names: alphabetical order
(`alpha`, `kilo`, `mike`, `zeta`) is not creation order here.

The `AGE` and `RESTARTS` columns can be read by eye, but in the exam a written command must work
again later, which is why the answer is a script and not a pasted list. `-o custom-columns`
makes the output easier to read without changing the sort:

```sh
kubectl get pods -A --sort-by=.metadata.creationTimestamp \
  -o custom-columns=NS:.metadata.namespace,NAME:.metadata.name,CREATED:.metadata.creationTimestamp
```

Quote the restart path: to the shell, `[0]` is a glob pattern.
