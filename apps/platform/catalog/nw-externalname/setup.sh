fresh_ns nw-externalname
fresh_ns nw-externalname-search
kubectl -n nw-externalname-search create deployment api --image=busybox:1.36 -- sh -c 'mkdir -p /www && echo "search results" > /www/index.html && httpd -f -p 80 -h /www' >/dev/null
kubectl -n nw-externalname-search expose deployment api --port=80 >/dev/null
kubectl -n nw-externalname run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-externalname-search rollout status deployment/api --timeout=120s >/dev/null
kubectl -n nw-externalname wait --for=condition=Ready pod/client --timeout=120s >/dev/null
