fresh_ns nw-dns-forward
kubectl apply -f - >/dev/null <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata:
  name: corp-dns
  namespace: nw-dns-forward
data:
  Corefile: |
    corp.internal:1053 {
        errors
        hosts {
            10.10.10.50 db.corp.internal
            10.10.10.51 ldap.corp.internal
        }
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: corp-dns
  namespace: nw-dns-forward
spec:
  replicas: 1
  selector:
    matchLabels:
      app: corp-dns
  template:
    metadata:
      labels:
        app: corp-dns
    spec:
      containers:
      - name: coredns
        image: registry.k8s.io/coredns/coredns:v1.14.6
        args: [-conf, /etc/coredns/Corefile]
        volumeMounts:
        - name: config
          mountPath: /etc/coredns
      volumes:
      - name: config
        configMap:
          name: corp-dns
---
apiVersion: v1
kind: Service
metadata:
  name: corp-dns
  namespace: nw-dns-forward
spec:
  selector:
    app: corp-dns
  ports:
  - name: dns
    port: 53
    targetPort: 1053
    protocol: UDP
  - name: dns-tcp
    port: 53
    targetPort: 1053
    protocol: TCP
YAML
kubectl -n nw-dns-forward run client --image=busybox:1.36 --command -- sleep 1d >/dev/null
kubectl -n nw-dns-forward rollout status deployment/corp-dns --timeout=120s >/dev/null
kubectl -n nw-dns-forward wait --for=condition=Ready pod/client --timeout=120s >/dev/null
