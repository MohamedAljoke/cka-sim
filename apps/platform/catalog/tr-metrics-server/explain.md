Applying the manifest as it is creates the Deployment, but its Pod never becomes Ready. Its logs
say why:

```text
x509: cannot validate certificate for 172.18.0.3 because it doesn't contain any IP SANs
```

metrics-server scrapes each kubelet over HTTPS and checks its serving certificate. Lab clusters
(kind, kubeadm without serving-cert approval) have self-signed kubelet certificates that name the
node, not its IP, so the check fails. The usual fix in a lab is one more argument for the
container:

```yaml
        - --kubelet-insecure-tls
```

```sh
kubectl apply -f /opt/course/tr-metrics-server/components.yaml
kubectl -n kube-system rollout status deployment metrics-server
kubectl top nodes
```

In production you'd keep TLS and fix the certificates instead: set `serverTLSBootstrap: true` in
the kubelet config, and approve the kubelets' serving CSRs.

The manifest also registers the APIService `v1beta1.metrics.k8s.io`, which tells the API server to
send `kubectl top`'s requests to metrics-server. `kubectl get apiservice v1beta1.metrics.k8s.io`
shows `False (MissingEndpoints)` while the Pod is not Ready.
