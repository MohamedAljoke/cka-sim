```sh
helm -n ar-helm-history history api
```

```text
REVISION  STATUS      CHART      APP VERSION  DESCRIPTION
1         superseded  web-1.0.0  1.26-alpine  Install complete
2         superseded  web-1.1.0  1.27-alpine  Upgrade complete
3         deployed    web-1.1.0  1.27-alpine  Upgrade complete
```

The `CHART` column is `<name>-<version>`, so revision 2 used version `1.1.0`.
`helm get metadata api --revision 2` prints it as a plain `VERSION:` line.

Every `helm get` subcommand takes `--revision`:

```sh
helm -n ar-helm-history get values api --revision 2 -o yaml
helm -n ar-helm-history get values api --revision 2 --all
helm -n ar-helm-history get manifest api --revision 2
```

Without `--all`, `get values` prints only what the user passed with `--set` or `-f`. With
`--all` it prints every value the templates saw, defaults included.

Revision 2 had both `replicaCount: 2` and `service.type: NodePort`, because it was upgraded with
`--reuse-values`. Revision 3 used `--set` without `--reuse-values`, so it started from the chart's
defaults again and kept only `replicaCount: 1`.
