`helm template` runs the chart's templates locally and prints the manifests. It talks to the
cluster for nothing and records no release:

```sh
helm template docs web --repo http://cka-sim-charts --version 1.0.0 -n ar-helm-template \
  > /opt/course/ar-helm-template/docs.yaml
kubectl -n ar-helm-template apply -f /opt/course/ar-helm-template/docs.yaml
```

The release name (`docs`) feeds the object names: this chart calls its Deployment
`<release>-<chart>`, so `docs-web`. `-n` sets `.Release.Namespace`, which charts use in
`metadata.namespace` or in URLs. Without it they'd render for `default`.

`helm ls` stays empty because the release record (Secret `sh.helm.release.v1.*`, label
`owner=helm`) is only written by `install` and `upgrade`. The objects still carry Helm's labels,
such as `app.kubernetes.io/managed-by: Helm`, because they're part of the templates. Labels don't
make a release.

The render also includes the chart's test Pod (`helm.sh/hook: test`). Add `--skip-tests` to leave
it out.
