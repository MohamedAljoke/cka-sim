pki=/etc/kubernetes/pki
answer() { grep -F "$1:" "$COURSE/ca.txt" 2>/dev/null | head -1 | cut -d: -f2- | tr -d ' \r'; }
signed() { local ca; ca=$(answer "$1"); [ -n "$ca" ] && [ -f "$ca" ] && openssl verify -CAfile "$ca" "$pki/$1" >/dev/null 2>&1; }

check 1 "apiserver.crt: its CA" signed apiserver.crt
check 1 "apiserver-etcd-client.crt: its CA" signed apiserver-etcd-client.crt
check 1 "front-proxy-client.crt: its CA" signed front-proxy-client.crt
check 1 "etcd/server.crt: its CA" signed etcd/server.crt
