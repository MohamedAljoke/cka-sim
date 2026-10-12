kubectl -n wl-blue-green patch service web -p '{"spec":{"selector":{"app":"web","version":"green"}}}'
kubectl -n wl-blue-green scale deployment web-blue --replicas=0
