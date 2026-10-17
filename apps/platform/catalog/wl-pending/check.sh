NS=wl-pending
field() { kubectl -n $NS get deployment "$1" -o jsonpath="$2"; }
ready() { eq 1 "$(field "$1" '{.status.readyReplicas}')"; }
# setup.sh creates both fresh, so generation 1 means nobody changed the spec yet.
edited() { [ "$(field "$1" '{.metadata.generation}')" -gt 1 ] 2>/dev/null; }
mebibytes() {
  case $1 in
    '') echo 0 ;;
    *Ki) echo $((${1%Ki} / 1024)) ;;
    *Mi) echo "${1%Mi}" ;;
    *Gi) echo $((${1%Gi} * 1024)) ;;
    *k) echo $((${1%k} * 1000 / 1048576)) ;;
    *M) echo $((${1%M} * 1000000 / 1048576)) ;;
    *G) echo $((${1%G} * 1000000000 / 1048576)) ;;
    *[0-9]) echo $(($1 / 1048576)) ;;
    *) echo 999999 ;;
  esac
}
small() { [ "$(mebibytes "$(field batch '{.spec.template.spec.containers[0].resources.requests.memory}')")" -le 256 ] 2>/dev/null; }
batch_ok() { small && ready batch; }
no_zone() { ready api && [ -z "$(kubectl get nodes -l zone -o name)" ]; }

if edited api; then
  check 2 "Deployment api runs 1 Ready Pod" wait_for 60 ready api
else
  echo "FAIL 2 Deployment api runs 1 Ready Pod"
fi
if edited batch; then
  check 2 "Deployment batch requests at most 256Mi and runs 1 Ready Pod" wait_for 60 batch_ok
else
  echo "FAIL 2 Deployment batch requests at most 256Mi and runs 1 Ready Pod"
fi
check 1 "api runs without a zone label on any node" no_zone
