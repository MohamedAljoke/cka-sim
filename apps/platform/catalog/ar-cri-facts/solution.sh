# kubectl get node cka-sim-control-plane -o wide -> CONTAINER-RUNTIME containerd://...
# grep containerRuntimeEndpoint /var/lib/kubelet/config.yaml
{
  echo "runtime: containerd"
  echo "socket: $(sed -n 's/^containerRuntimeEndpoint: *//p' /var/lib/kubelet/config.yaml)"
  echo "kube-proxy: $(crictl ps -q --name '^kube-proxy$' | head -1)"
} > /opt/course/ar-cri-facts/runtime.txt
