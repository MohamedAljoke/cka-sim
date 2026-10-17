answer() { tr -d ' \r' < "$COURSE/components.txt" 2>/dev/null | tr 'A-Z' 'a-z' | sed -n "s/^$1:\(.*\)$/\1/p" | head -1; }
is() { eq "$2" "$(answer "$1")"; }
node_agent() { is kubelet systemd; }
static() { is kube-apiserver static-pod && is kube-scheduler static-pod && is kube-controller-manager static-pod && is etcd static-pod; }
addons() { is coredns deployment && is kube-proxy daemonset; }

check 1 "kubelet runs as a systemd service" node_agent
check 2 "kube-apiserver, kube-scheduler, kube-controller-manager and etcd are static Pods" static
check 2 "coredns is a Deployment and kube-proxy a DaemonSet" addons
