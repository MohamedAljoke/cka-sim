Helm guards each release with its latest revision's status. While that is `pending-install`,
`pending-upgrade` or `pending-rollback`, it refuses every new operation, because another Helm
process might still be working. When that process was killed, nothing ever clears the status:

```sh
helm -n ar-helm-stuck history billing
```

```text
REVISION  STATUS           CHART      DESCRIPTION
1         deployed         web-1.0.0  Install complete
2         pending-upgrade  web-1.1.0  Preparing upgrade
```

A rollback is allowed in this state and writes a new, clean revision:

```sh
helm -n ar-helm-stuck rollback billing 1
```

Revision 3 (`Rollback to 1`) is `deployed`, and upgrades work again. It also undoes whatever the
half-done upgrade already changed: here the Deployment was switched to an image tag that doesn't
exist.

Other ways out you'll find online:

- Deleting the pending revision's Secret `sh.helm.release.v1.billing.v2` makes revision 1 current
  again, but leaves the half-applied changes running.
- `helm uninstall` loses the history and briefly deletes the app.
