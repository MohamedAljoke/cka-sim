The kubelet reads every file in its `staticPodPath` (here `/etc/kubernetes/manifests`) every few
seconds. A file it can't parse is skipped. Nothing is created, not even a failed container, so
`kubectl` and `crictl` show nothing at all. The only trace is in the kubelet's own log:

```sh
grep staticPodPath /var/lib/kubelet/config.yaml
journalctl -u kubelet --no-pager | grep -i node-exporter | tail -3
#   "Could not process manifest file" ... error converting YAML to JSON: yaml: line 10: mapping values are not allowed in this context
cat -A /etc/kubernetes/manifests/node-exporter.yaml | sed -n 8,11p     # show the indentation
vim /etc/kubernetes/manifests/node-exporter.yaml
```

The line number points at the problem: `command:` is indented one space more than `image:`, so
YAML reads it as part of the image value. Validate a manifest before dropping it in with
`kubectl apply --dry-run=client -f file.yaml`, or `kubectl create --dry-run=server` against a
real API server.

Once the file parses, the kubelet starts the Pod and creates its mirror,
`node-exporter-cka-sim-worker2`, in the namespace from the manifest.
