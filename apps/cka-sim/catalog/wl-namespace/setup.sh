# galaxy is the candidate's to create, so it carries no tidy label: clear it here.
kubectl delete namespace galaxy --ignore-not-found --wait=true --timeout=120s >/dev/null
fresh_course
planets=(planet-a planet-b planet-c)
home=${planets[RANDOM % 3]}
for ns in "${planets[@]}"; do
  fresh_ns "$ns"
  if [ "$ns" = "$home" ]; then
    kubectl -n "$ns" run hermes --image=nginx:1.27 >/dev/null
  else
    kubectl -n "$ns" run ares --image=nginx:1.27 >/dev/null
  fi
done
