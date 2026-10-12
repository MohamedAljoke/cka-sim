E() { etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/server.crt --key=/etc/kubernetes/pki/etcd/server.key "$@"; }
answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/etcd.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
version_ok() { eq "$(E endpoint status -w json | jq -r '.[0].Status.version')" "$(answer version | sed 's/^v//')"; }
datadir_ok() { eq "$(grep -o -- '--data-dir=[^ ]*' /etc/kubernetes/manifests/etcd.yaml | cut -d= -f2)" "$(answer data-dir | sed 's|/$||')"; }
members_ok() { eq "$(E member list -w json | jq -r '[.members[].name] | sort | join(",")')" "$(answer members | tr ',' '\n' | sort | paste -sd,)"; }

check 1 "version is the etcd server version" version_ok
check 1 "data-dir is etcd's data directory" datadir_ok
check 1 "members lists every etcd member" members_ok
