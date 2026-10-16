The **downward API** lets a container learn about its own Pod without calling the API server. It
comes in two forms:

- **env vars** (`valueFrom.fieldRef` / `resourceFieldRef`): fixed at container start; good for the
  Pod's name, namespace, IP or node name
- **volume files**: labels and annotations are only available this way, and the files **update**
  when the labels change

```yaml
volumes:
- name: podinfo
  downwardAPI:
    items:
    - {path: labels,      fieldRef: {fieldPath: metadata.labels}}
    - {path: annotations, fieldRef: {fieldPath: metadata.annotations}}
    - path: mem_limit
      resourceFieldRef: {containerName: meta, resource: limits.memory, divisor: 1Mi}
```

`divisor` sets the unit: without it the memory limit is written in bytes (`67108864`). For CPU, use
`divisor: 1m` to get millicores. Labels are written one per line as `key="value"`.

Apps use this to tag their logs and metrics with the Pod's identity, or to size their heap from
the container's memory limit.
