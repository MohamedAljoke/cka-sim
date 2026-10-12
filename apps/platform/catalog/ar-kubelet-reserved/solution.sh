cfg=/var/lib/kubelet/config.yaml
sed -i 's/^evictionHard:$/evictionHard:\n  memory.available: 300Mi/' "$cfg"
printf 'systemReserved:\n  cpu: 500m\n  memory: 512Mi\n' >> "$cfg"
systemctl restart kubelet
until kubectl get node cka-sim-worker2 -o jsonpath='{.status.allocatable.cpu}' | grep -q m; do sleep 2; done
kubectl describe node cka-sim-worker2 | grep -A6 Allocatable
