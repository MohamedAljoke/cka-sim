readyz_ok() { grep -q '^\[+\]etcd ok' "$COURSE/readyz.txt" 2>/dev/null && grep -q 'readyz check passed' "$COURSE/readyz.txt"; }
kubelet_ok() { eq ok "$(tr -d ' \r\n' < "$COURSE/kubelet.txt" 2>/dev/null)"; }

check 2 "readyz.txt holds the verbose readyz report, etcd included" readyz_ok
check 2 "kubelet.txt holds the kubelet's healthz answer" kubelet_ok
