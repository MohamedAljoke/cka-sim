conf=$(kubectl -n nw-dns-config exec client -- cat /etc/resolv.conf)
{
  echo "search: $(echo "$conf" | sed -n 's/^search //p')"
  echo "ndots: $(echo "$conf" | grep -o 'ndots:[0-9]*' | cut -d: -f2)"
} > /opt/course/nw-dns-config/client.txt
kubectl apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: custom-dns
  namespace: nw-dns-config
spec:
  dnsPolicy: None
  dnsConfig:
    nameservers: [10.96.0.10]
    searches: [nw-dns-config.svc.cluster.local]
    options:
    - name: ndots
      value: "2"
  containers:
  - name: app
    image: busybox:1.36
    command: [sleep, 1d]
YAML
kubectl -n nw-dns-config wait --for=condition=Ready pod/custom-dns --timeout=60s
