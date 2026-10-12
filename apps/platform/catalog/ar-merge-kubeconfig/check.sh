KC="$COURSE/merged.kubeconfig"
K() { kubectl --kubeconfig "$KC" --request-timeout=5s "$@"; }
both() { eq "team-a team-b" "$(K config get-contexts -o name 2>/dev/null | sort | xargs)"; }
embedded() { both && [ -z "$(K config view --raw -o jsonpath='{.users[*].user.client-certificate}{.clusters[*].cluster.certificate-authority}')" ] && [ -n "$(K config view --raw -o jsonpath='{.users[?(@.name=="admin-a")].user.client-certificate-data}')" ]; }
current() { eq team-b "$(K config current-context 2>/dev/null)" && eq ar-merge-b "$(K config view -o jsonpath='{.contexts[?(@.name=="team-b")].context.namespace}')"; }
works() { current && eq b "$(K get configmap marker -o jsonpath='{.data.team}' 2>/dev/null)" && K --context team-a get namespace default >/dev/null; }

check 2 "merged.kubeconfig holds contexts team-a and team-b" both
check 1 "Certificates are embedded in it" embedded
check 1 "team-b is current, with namespace ar-merge-b" current
check 1 "Both contexts work, and team-b reads ar-merge-b by default" works
