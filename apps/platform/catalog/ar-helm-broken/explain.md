`helm ls` shows only deployed releases. Add `-a` to see all, or filter by state:

```sh
helm -n ar-helm-broken ls -a
helm -n ar-helm-broken ls --failed
helm -n ar-helm-broken status cart
helm -n ar-helm-broken uninstall cart
```

`cart` was installed with `--wait` and an image tag that doesn't exist, so its Pods never became
Ready and Helm marked the release `failed`. A failed release still owns its objects (the
Deployment `cart-web` is there, stuck in `ImagePullBackOff`), and `uninstall` removes them with the
release.

Other states worth knowing:

- `pending-install` and `pending-upgrade`: a Helm process died mid-operation.
- `uninstalling`: an uninstall is in progress.
- `superseded`: an older revision.

The chart's defaults come from `helm show values`, which works without installing anything:

```sh
helm show values web --repo http://cka-sim-charts --version 1.1.0 > /opt/course/ar-helm-broken/values.yaml
```

`helm show values cka/web --version 1.1.0` is the same, once the repository is added as `cka`.
