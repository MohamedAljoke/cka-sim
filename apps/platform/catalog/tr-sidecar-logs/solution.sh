kubectl -n tr-sidecar-logs patch deployment orders --type=json -p '[{"op":"add","path":"/spec/template/spec/containers/-","value":
  {"name":"log-streamer","image":"busybox:1.36","command":["sh","-c","tail -n+1 -F /var/log/orders/orders.log"],
   "volumeMounts":[{"name":"logs","mountPath":"/var/log/orders","readOnly":true}]}}]'
kubectl -n tr-sidecar-logs rollout status deployment orders --timeout=120s
