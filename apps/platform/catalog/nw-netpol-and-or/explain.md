The policy had two entries under `from`, and entries are **ORed**:

```yaml
  - from:
    - namespaceSelector:          # any Pod in a team=monitoring Namespace
        matchLabels:
          team: monitoring
    - podSelector:                # OR any app=prometheus Pod in this Namespace
        matchLabels:
          app: prometheus
```

Put both selectors in **one** entry (one leading `-`) and they're **ANDed**:

```yaml
  - from:
    - namespaceSelector:
        matchLabels:
          team: monitoring
      podSelector:
        matchLabels:
          app: prometheus
```

The only difference is one `-`, so watch the indentation. `kubectl describe networkpolicy` shows
which you have: ANDed selectors print together on one line, as "NamespaceSelector: team=monitoring
PodSelector: app=prometheus".

```bash
kubectl -n nw-netpol-and-or edit networkpolicy api-from-monitoring
```
