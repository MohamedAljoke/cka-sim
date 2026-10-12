```yaml
  ingress:
  - ports:
    - protocol: TCP
      port: 8000
      endPort: 8010
    - protocol: TCP
      port: metrics
```

- **`endPort` turns `port` into a range,** inclusive at both ends. It only works with a numeric
  `port`, and `endPort` must be greater than or equal to `port`.
- **A named `port` matches the container port with that name** on each selected Pod, so the policy
  keeps working if the number changes. A named port can't be combined with `endPort`.
- **A rule with `ports` but no `from` admits any source,** but only on those ports.

`kubectl explain networkpolicy.spec.ingress.ports` documents both fields.
