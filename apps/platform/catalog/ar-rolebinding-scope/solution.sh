kubectl -n ar-scope-dev create rolebinding ci-pod-reader --clusterrole=pod-reader --serviceaccount=ar-scope-dev:ci
kubectl auth can-i list pods -n ar-scope-dev --as=system:serviceaccount:ar-scope-dev:ci
