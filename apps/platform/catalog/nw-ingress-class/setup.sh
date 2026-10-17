kubectl annotate ingressclass nginx ingressclass.kubernetes.io/is-default-class- >/dev/null 2>&1
fresh_ns nw-ingress-class
fresh_course
web_app nw-ingress-class legacy
cat > "$COURSE/legacy.yaml" <<'YAML'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: legacy
  namespace: nw-ingress-class
spec:
  rules:
  - host: legacy.nw-ingress-class.test
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: legacy
            port:
              number: 80
YAML
kubectl apply -f "$COURSE/legacy.yaml" >/dev/null
