A tag is a movable name; a digest is the hash of the image manifest and can't point anywhere else.
The kubelet records the digest it actually pulled in each container's `imageID`:

```sh
kubectl -n wl-image-digest get pods -l app=web \
  -o jsonpath='{.items[0].status.containerStatuses[0].imageID}'
# cka-sim-registry:5000/team/web@sha256:91fc…
kubectl -n wl-image-digest set image deployment web web=cka-sim-registry:5000/team/web@sha256:91fc…
```

`crictl images --digests` on the node shows the same. A reference can carry both forms,
`team/web:1.1@sha256:…`; the digest wins and the tag is only a label for humans.

With a digest, `imagePullPolicy` no longer decides *what* runs, only *whether* the kubelet asks the
registry again. `IfNotPresent` skips the pull when the node has that exact image, so Pods still
start when the registry is down. `Always` with a tag is the opposite trade: every start may get a
different image.

`imageID` holds the registry digest only for images pulled from a registry. Images loaded with
`ctr import` or `kind load` report a local ID that no registry can serve.
