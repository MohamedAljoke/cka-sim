```yaml
  rules:
  - matches:
    - headers:
      - name: x-version
        value: v2
    backendRefs:
    - name: v2
      port: 80
  - backendRefs:
    - name: v1
      port: 80
```

- **Conditions inside one `matches` entry are ANDed** (path, headers, query params, method). Separate
  entries in `matches` are ORed.
- **Header names are case-insensitive, values are exact** unless you set `type: RegularExpression`.
- **The more specific rule wins.** A rule that matches a header beats one that matches only the
  path, so the order you write them in doesn't matter.

```bash
curl -H 'Host: app.nw-gateway-header.test' -H 'x-version: v2' http://10.96.80.80/   # v2 /
curl -H 'Host: app.nw-gateway-header.test' http://10.96.80.80/                      # v1 /
```

To change headers on the way in instead, use a `RequestHeaderModifier` filter
(`filters: [{type: RequestHeaderModifier, requestHeaderModifier: {set: [{name: x-env, value: prod}]}}]`).
