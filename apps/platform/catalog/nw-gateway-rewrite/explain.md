```yaml
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /legacy
    filters:
    - type: URLRewrite
      urlRewrite:
        path:
          type: ReplacePrefixMatch
          replacePrefixMatch: /
    backendRefs:
    - name: app
      port: 80
```

- **`ReplacePrefixMatch` replaces only the part the `PathPrefix` matched,** so `/legacy/items` becomes
  `/items`. It needs a `PathPrefix` match in the same rule.
- **`ReplaceFullPath`** would send every request to one fixed path instead.
- **`urlRewrite.hostname`** rewrites the `Host` header the backend sees.

```bash
curl -H 'Host: legacy.nw-gateway-rewrite.test' http://10.96.80.80/legacy/items   # app /items
```

A rewrite changes the request before it's forwarded, and the client never sees it. A `RequestRedirect`
filter instead tells the client to request a new URL.
