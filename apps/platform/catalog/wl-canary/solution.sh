kubectl -n wl-canary get deployment web-stable -o yaml |
  sed -e 's/name: web-stable/name: web-canary/' -e 's/track: stable/track: canary/' -e 's/value: stable/value: canary/' -e 's/replicas: 3/replicas: 1/' |
  grep -v -E '^\s*(uid|resourceVersion|creationTimestamp|generation):' | kubectl apply -f -
kubectl -n wl-canary rollout status deployment web-canary --timeout=120s
