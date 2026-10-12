```bash
kubectl -n nw-ingress-tls create secret tls secure-tls \
  --cert=/opt/course/nw-ingress-tls/tls.crt --key=/opt/course/nw-ingress-tls/tls.key
kubectl -n nw-ingress-tls edit ingress secure
```

```yaml
spec:
  tls:
  - hosts:
    - secure.nw-ingress-tls.test
    secretName: secure-tls
```

- **The Secret must be type `kubernetes.io/tls`,** with keys `tls.crt` and `tls.key`. `kubectl create
  secret tls` builds exactly that. A generic Secret is ignored.
- **The Secret lives in the Ingress's Namespace.** An Ingress can't reference one elsewhere.
- **The `hosts` under `tls` must match the rule's host.** The controller picks the certificate by the
  name the client sends in SNI.

Test it, and look at which certificate comes back:

```bash
curl -k --resolve secure.nw-ingress-tls.test:443:10.96.80.80 https://secure.nw-ingress-tls.test/
openssl s_client -connect 10.96.80.80:443 -servername secure.nw-ingress-tls.test </dev/null 2>/dev/null \
  | openssl x509 -noout -subject
```

Without the `tls` section you'd get the controller's built-in default certificate instead.
