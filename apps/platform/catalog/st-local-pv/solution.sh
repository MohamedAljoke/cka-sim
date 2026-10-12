# kubectl -n st-local-pv describe pod app -> 0/3 nodes are available: ... 1 node(s) had volume node affinity conflict ...
# kubectl get pv local-data -o yaml        -> nodeAffinity: kubernetes.io/hostname In [cka-sim-worker2]
f=/opt/course/st-local-pv/app.yaml
sed -i 's/kubernetes.io\/hostname: cka-sim-worker$/kubernetes.io\/hostname: cka-sim-worker2/' "$f"
kubectl replace --force -f "$f"
kubectl -n st-local-pv wait --for=condition=Ready pod/app --timeout=90s
