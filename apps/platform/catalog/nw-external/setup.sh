fresh_ns nw-external
fresh_ns nw-external-legacy
fresh_course
kubectl -n nw-external-legacy run legacy --image=busybox:1.36 --command -- \
  sh -c 'mkdir -p /www && echo "legacy db api" > /www/index.html && httpd -f -p 8080 -h /www' >/dev/null
kubectl -n nw-external run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-external-legacy wait --for=condition=Ready pod/legacy --timeout=120s >/dev/null
kubectl -n nw-external wait --for=condition=Ready pod/client --timeout=120s >/dev/null
kubectl -n nw-external-legacy get pod legacy -o jsonpath='{.status.podIP}' > "$COURSE/backend-ip"
