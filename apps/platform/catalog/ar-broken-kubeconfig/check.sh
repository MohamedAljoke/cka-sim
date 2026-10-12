KC="$COURSE/dev.kubeconfig"
K() { kubectl --kubeconfig "$KC" --request-timeout=5s "$@"; }
works() { K get nodes >/dev/null 2>&1; }
current() { works && eq dev "$(K config current-context)"; }
same_creds() { works && eq "$(kubectl --kubeconfig /etc/kubernetes/admin.conf config view --raw --minify -o jsonpath='{.users[0].user.client-certificate-data}')" "$(K config view --raw --minify -o jsonpath='{.users[0].user.client-certificate-data}')"; }

check 3 "kubectl --kubeconfig dev.kubeconfig get nodes works" works
check 1 "Context dev is the current context" current
check 1 "It still uses the original credentials" same_creds
