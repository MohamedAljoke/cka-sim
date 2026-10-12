answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/auth.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
flag() { grep -o -- "--$1=[^ ]*" /etc/kubernetes/manifests/kube-apiserver.yaml | head -1 | cut -d= -f2; }
anonymous() { local v; v=$(flag anonymous-auth); echo "${v:-true}"; }

check 1 "client-ca-file is the API server's client CA" eq "$(flag client-ca-file)" "$(answer client-ca-file)"
check 1 "anonymous-auth is right (true unless the flag disables it)" eq "$(anonymous)" "$(answer anonymous-auth)"
check 1 "authorization-mode lists the authorizers in order" eq "$(flag authorization-mode)" "$(answer authorization-mode)"
