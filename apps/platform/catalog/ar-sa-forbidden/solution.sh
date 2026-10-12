# kubectl -n ar-sa get rolebindings -o wide           -> reporter-pod-reader binds Role pod-reader to reporter
# kubectl -n ar-sa get role pod-reader -o yaml        -> verbs get, watch: list is missing
kubectl -n ar-sa patch role pod-reader --type=json -p '[{"op":"add","path":"/rules/0/verbs/-","value":"list"}]'
kubectl auth can-i list pods -n ar-sa --as=system:serviceaccount:ar-sa:reporter
