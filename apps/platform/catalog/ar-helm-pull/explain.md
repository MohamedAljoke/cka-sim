```sh
helm pull web --repo http://cka-sim-charts --version 1.1.0 --untar --untardir /opt/course/ar-helm-pull
vi /opt/course/ar-helm-pull/web/values.yaml        # replicaCount: 4
helm -n ar-helm-local install local /opt/course/ar-helm-pull/web --create-namespace
```

`helm pull` downloads the `.tgz`. `--untar` unpacks it, and `--untardir` chooses where; the
folder is named after the chart. With the repository added (`helm repo add cka …`), it's
`helm pull cka/web --version 1.1.0 --untar`.

`helm install` takes a chart reference in several forms:

- `cka/web`, from an added repository
- `web --repo <url>`
- a `.tgz` file
- a folder, like here

From a folder, the changed `values.yaml` *is* the chart's default, so `helm get values local`
shows no user-supplied values and `helm get values local --all` shows `replicaCount: 4`.

`--create-namespace` makes Helm create `ar-helm-local` first; without it the install fails with
`namespaces "ar-helm-local" not found`. Helm doesn't own that Namespace: `helm uninstall` leaves
it behind.
