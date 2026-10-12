cidr=$(cat /opt/course/nw-netpol-ipblock/cidr)
ip=$(cat /opt/course/nw-netpol-ipblock/except)
kubectl apply -f - <<YAML
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: web-ipblock
  namespace: nw-netpol-ipblock
spec:
  podSelector:
    matchLabels:
      app: web
  policyTypes: [Ingress]
  ingress:
  - from:
    - ipBlock:
        cidr: $cidr
        except:
        - $ip/32
    ports:
    - protocol: TCP
      port: 80
YAML
sleep 5
