# kubectl -n shop get endpointslices -l kubernetes.io/service-name=web -> no addresses: the selector matches no Pod
# kubectl -n shop get pods --show-labels  -> app=web,tier=frontend; the Service selects app=webapp
# kubectl -n shop describe service web    -> TargetPort 8080; nginx listens on 80
kubectl -n shop patch service web --type=json -p '[
  {"op":"replace","path":"/spec/selector","value":{"app":"web"}},
  {"op":"replace","path":"/spec/ports/0/targetPort","value":80}]'
