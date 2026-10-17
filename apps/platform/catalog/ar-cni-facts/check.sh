answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/cni.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
plugin_ok() { [[ "$(answer plugin | tr 'A-Z' 'a-z')" == kindnet* ]]; }
config_ok() { eq "$(ls /etc/cni/net.d/* | head -1)" "$(answer config)"; }
cidr_ok() { eq "$(grep -o -- '--cluster-cidr=[^ ]*' /etc/kubernetes/manifests/kube-controller-manager.yaml | cut -d= -f2)" "$(answer podCIDR)"; }

check 1 "plugin is kindnet" plugin_ok
check 1 "config is the CNI config file in /etc/cni/net.d" config_ok
check 1 "podCIDR is the cluster-wide Pod CIDR" cidr_ok
