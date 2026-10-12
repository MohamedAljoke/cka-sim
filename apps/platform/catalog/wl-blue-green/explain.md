**Blue/green** runs the new version fully next to the old one, then moves all traffic at once by
changing the Service's **selector**. The switch is atomic: the EndpointSlice controller swaps the
endpoints in one step, and rolling back means patching the selector back.

```sh
kubectl -n wl-blue-green get pods --show-labels
kubectl -n wl-blue-green get deploy web-green                        # make sure green is Ready first
kubectl -n wl-blue-green patch svc web -p '{"spec":{"selector":{"app":"web","version":"green"}}}'
kubectl -n wl-blue-green exec client -- wget -qO- http://web         # green
kubectl -n wl-blue-green scale deploy web-blue --replicas=0          # keep it around for a fast rollback
```

Compared with a rolling update, you need twice the capacity during the switch, but you never mix
versions and the rollback is instant. A patch merges the selector map, so give the full map; keys
you leave out stay as they were. Check that the target Pods are Ready **before** switching, or you
get a gap with no backends.
