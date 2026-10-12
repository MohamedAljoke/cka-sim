fresh_ns wl-secret-decode
fresh_course
kubectl -n wl-secret-decode create secret generic api-token \
  --from-literal=token="tk-$(head -c 12 /dev/urandom | od -An -tx1 | tr -d ' \n')" --from-literal=issuer=ci >/dev/null
