ip=$(cat /opt/course/nw-external/backend-ip)
kubectl apply -f - <<YAML
apiVersion: v1
kind: Service
metadata:
  name: legacy-db
  namespace: nw-external
spec:
  ports:
  - name: http
    port: 80
    targetPort: 8080
---
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: legacy-db-1
  namespace: nw-external
  labels:
    kubernetes.io/service-name: legacy-db
addressType: IPv4
ports:
- name: http
  port: 8080
  protocol: TCP
endpoints:
- addresses: ["$ip"]
YAML
