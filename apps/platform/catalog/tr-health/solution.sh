kubectl get --raw='/readyz?verbose' > /opt/course/tr-health/readyz.txt
kubectl get --raw=/api/v1/nodes/cka-sim-worker2/proxy/healthz > /opt/course/tr-health/kubelet.txt
