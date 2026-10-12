A Service has two ports. `port` is what clients connect to on the Service's IP, and
`targetPort` is where traffic goes in the Pod. They don't have to match, and that's the point:
clients use a stable port while the app listens wherever it likes.

```sh
kubectl -n nw-clusterip expose deployment api --port=80 --target-port=8080
kubectl -n nw-clusterip get svc,endpointslices
kubectl -n nw-clusterip exec client -- wget -qO- http://api.nw-clusterip
```

`kubectl expose` copies the Deployment's selector into the Service, which avoids the most common
mistake, a selector that matches no Pods. Without `--target-port`, it uses the same value as
`--port`, so here traffic would go to port 80 in the Pod, where nothing listens.

`kubectl create service clusterip api --tcp=80:8080` also works, but it selects `app=api`
because of the name. That only works when the labels happen to match.
