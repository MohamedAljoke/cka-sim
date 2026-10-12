By default a Job retries every failed Pod, with growing back-off, until `backoffLimit`. That wastes
minutes when the failure is permanent. A **podFailurePolicy** looks at *why* a Pod failed and picks
an action per rule:

| Action | Effect |
|---|---|
| `FailJob` | mark the whole Job failed at once |
| `Ignore` | don't count this failure against `backoffLimit` (e.g. a Pod evicted by a drain) |
| `Count` | the default behaviour |
| `FailIndex` | for indexed Jobs, give up on that index only |

```yaml
spec:
  backoffLimit: 6
  podFailurePolicy:
    rules:
    - action: FailJob
      onExitCodes: {containerName: importer, operator: In, values: [42]}
    - action: Ignore
      onPodConditions: [{type: DisruptionTarget}]      # evictions don't count
  template:
    spec:
      restartPolicy: Never        # required: with OnFailure the kubelet restarts the container instead
```

```sh
kubectl -n wl-pod-failure-policy get job importer -o jsonpath='{.status.conditions[?(@.type=="Failed")]}'
#   reason PodFailurePolicy, message: Container importer for pod ... failed with exit code 42 matching FailJob rule
```

Exit codes are the app's contract with the platform: reserve a few for "don't bother retrying".
