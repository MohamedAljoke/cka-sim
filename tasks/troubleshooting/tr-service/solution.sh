# k -n shop get ep web                 -> <none>: the selector matches no Pod
# k -n shop get pods --show-labels      -> app=web,tier=frontend ; the Service selects app=webapp
# k -n shop describe svc web            -> TargetPort 8080 ; nginx listens on 80
k -n shop patch service web --type=json -p '[
  {"op":"replace","path":"/spec/selector","value":{"app":"web"}},
  {"op":"replace","path":"/spec/ports/0/targetPort","value":80}]' >/dev/null
