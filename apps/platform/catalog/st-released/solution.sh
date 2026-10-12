# kubectl get pv pv-archive                      -> Released, CLAIM st-released/archive (the old claim)
kubectl patch pv pv-archive --type=json -p '[{"op":"remove","path":"/spec/claimRef"}]'
kubectl -n st-released wait --for=jsonpath='{.status.phase}'=Bound pvc/archive-v2 --timeout=60s
kubectl -n st-released wait --for=condition=Ready pod/reader --timeout=90s
