#!/usr/bin/env bash
# The pool of warm exam VMs. Each one is created, builds its cluster, is marked pool=ready and suspended;
# the cka-sim server resumes one per exam and destroys it afterwards.
#
#   pool.sh fill N   add N warm VMs (in parallel, ~2 minutes)
#   pool.sh list     every VM in the pool app and its pool state
#   pool.sh empty    destroy every VM in the pool app, including ones in use
set -euo pipefail

APP=${FLY_LABS_APP:-cka-sim-labs-maljoke}
REGION=${FLY_REGION:-gru}
IMAGE=registry.fly.io/$APP:lab
API=https://api.machines.dev/v1/apps/$APP

api() {
  local method=$1 path=$2
  shift 2
  curl -fsS -X "$method" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" "$API$path" "$@"
}

add_one() {
  local id
  id=$(fly machine run "$IMAGE" -a "$APP" -r "$REGION" \
    --vm-cpu-kind performance --vm-cpus 2 --vm-memory 4096 --rootfs-size 20 \
    --restart no --metadata pool=building --detach 2>&1 | sed -n 's/.*Machine ID: *\([0-9a-f]*\).*/\1/p')
  if [ -z "$id" ]; then
    echo "could not create a VM" >&2
    return 1
  fi
  echo "$id: created, building the cluster"
  if api GET "/machines/$id/wait?state=started&timeout=60" >/dev/null &&
    build "$id" &&
    api POST "/machines/$id/metadata/pool" -d '{"value":"ready"}' >/dev/null &&
    api POST "/machines/$id/suspend" >/dev/null &&
    api GET "/machines/$id/wait?state=suspended&timeout=60" >/dev/null; then
    echo "$id: warm"
  else
    echo "$id: failed, destroying it" >&2
    api DELETE "/machines/$id?force=true" >/dev/null || true
    return 1
  fi
}

# ssh into a just-started VM can fail for a few seconds.
build() {
  local id=$1 try
  for try in 1 2 3 4 5 6; do
    if fly ssh console -a "$APP" --machine "$id" -C lab-build >"/tmp/cka-sim-pool-$id.log" 2>&1; then
      return 0
    fi
    grep -q "cluster" "/tmp/cka-sim-pool-$id.log" && break
    sleep 5
  done
  echo "$id: build failed, see /tmp/cka-sim-pool-$id.log" >&2
  return 1
}

TOKEN=$(fly auth token 2>/dev/null | tail -1)
case "${1:-}" in
fill)
  n=${2:?how many VMs?}
  for _ in $(seq "$n"); do add_one & done
  wait
  ;;
list)
  fly machine list -a "$APP" --json | jq -r '.[] | [.id, .state, (.config.metadata.pool // "-")] | @tsv'
  ;;
empty)
  for id in $(fly machine list -a "$APP" --json | jq -r '.[].id'); do
    api DELETE "/machines/$id?force=true" >/dev/null && echo "$id: destroyed"
  done
  ;;
*)
  sed -n '2,8p' "$0"
  exit 2
  ;;
esac
