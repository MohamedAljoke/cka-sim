answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/proxy.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
mode() { kubectl -n kube-system get configmap kube-proxy -o jsonpath='{.data.config\.conf}' | sed -n 's/^mode: *"\{0,1\}\([a-z]*\)"\{0,1\}$/\1/p'; }
chain() { iptables-save -t nat | grep 'nw-kube-proxy-mode/api' | grep -o 'KUBE-SVC-[A-Z0-9]*' | head -1; }
seps() { iptables-save -t nat | grep "^-A $(chain) " | grep -o 'KUBE-SEP-[A-Z0-9]*' | sort -u | wc -l; }

check 1 "mode is kube-proxy's mode" eq "$(mode)" "$(answer mode)"
check 2 "chain is the KUBE-SVC chain of Service api" eq "$(chain)" "$(answer chain)"
check 1 "endpoints is the number of KUBE-SEP chains" eq "$(seps)" "$(answer endpoints)"
