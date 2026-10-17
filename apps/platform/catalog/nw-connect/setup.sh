fresh_ns nw-connect
fresh_course
kubectl -n nw-connect run backend --image=busybox:1.36 --labels=app=backend --port=8080 --command -- \
  sh -c 'mkdir -p /www && echo "backend $(hostname) says hello" > /www/index.html && httpd -f -p 8080 -h /www' >/dev/null
kubectl -n nw-connect expose pod backend --port=80 --target-port=8080 >/dev/null
kubectl -n nw-connect wait --for=condition=Ready pod/backend --timeout=120s >/dev/null
