```sh
helm repo add cka http://cka-sim-charts
helm search repo cka/web --versions
helm -n ar-helm-install install shop cka/web --version 1.0.0 --set replicaCount=3
helm -n ar-helm-install upgrade shop cka/web --version 1.1.0 --reuse-values
helm -n ar-helm-install history shop
helm -n ar-helm-install rollback shop 1
```

Each install, upgrade and rollback is a new **revision**, stored as a Secret
`sh.helm.release.v1.shop.v<N>` in the release's Namespace. A rollback doesn't delete revision 2;
it creates revision 3 with revision 1's chart and values, so `history` shows
`Rollback to 1`.

The upgrade is where values get lost. Without any `--set` or `-f`, `helm upgrade` reuses the last
values. But as soon as you pass one, it starts again from the chart's defaults. `--reuse-values`
(or repeating `--set replicaCount=3`) keeps the 3 Pods either way. Chart 1.1.0 defaults to 2.

`-n` matters for every command: releases are per Namespace, and `helm ls` without `-n` or `-A`
only shows the current one.
