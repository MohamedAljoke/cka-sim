# Without nodeTaintsPolicy: Honor the tainted control plane counts as an empty domain, and
# DoNotSchedule then leaves 2 Pods Pending.
kubectl -n wl-spread patch deployment web --type=json -p '[
  {"op":"remove","path":"/spec/template/spec/affinity"},
  {"op":"add","path":"/spec/template/spec/topologySpreadConstraints","value":[
    {"maxSkew":1,"topologyKey":"kubernetes.io/hostname","whenUnsatisfiable":"DoNotSchedule",
     "nodeTaintsPolicy":"Honor","labelSelector":{"matchLabels":{"app":"web"}},"matchLabelKeys":["pod-template-hash"]}]}]'
kubectl -n wl-spread rollout status deployment/web --timeout=120s
