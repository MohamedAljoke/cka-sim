{
  echo "pods: $(grep -o -- '--cluster-cidr=[^ ]*' /etc/kubernetes/manifests/kube-controller-manager.yaml | cut -d= -f2)"
  echo "services: $(grep -o -- '--service-cluster-ip-range=[^ ]*' /etc/kubernetes/manifests/kube-apiserver.yaml | cut -d= -f2)"
  echo "cka-sim-worker: $(kubectl get node cka-sim-worker -o jsonpath='{.spec.podCIDR}')"
} > /opt/course/nw-cidrs/cidrs.txt
