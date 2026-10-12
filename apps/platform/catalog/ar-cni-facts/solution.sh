# ls /etc/cni/net.d/                         -> 10-kindnet.conflist
# kubectl -n kube-system get ds              -> kindnet
# grep cluster-cidr /etc/kubernetes/manifests/kube-controller-manager.yaml
{
  echo "plugin: kindnet"
  echo "config: $(ls /etc/cni/net.d/* | head -1)"
  echo "podCIDR: $(grep -o -- '--cluster-cidr=[^ ]*' /etc/kubernetes/manifests/kube-controller-manager.yaml | cut -d= -f2)"
} > /opt/course/ar-cni-facts/cni.txt
