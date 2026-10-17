answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/resources.txt" 2>/dev/null | head -1 | tr -d '\r' | xargs; }
truth() {
  kubectl api-resources --no-headers 2>/dev/null | awk -v r="$1" '$1==r { gv=(NF==5)?$3:$2; ns=(NF==5)?$4:$3; print gv, (ns=="true"?"namespaced":"cluster"); exit }'
}
right() { [ -n "$(truth "$1")" ] && eq "$(truth "$1")" "$(answer "$1")"; }
all() { for r in "$@"; do right "$r" || return 1; done; }

check 1 "deployments and leases are right" all deployments leases
check 1 "priorityclasses and storageclasses are right" all priorityclasses storageclasses
check 1 "events is right" all events
check 1 "csinodes is right" all csinodes
