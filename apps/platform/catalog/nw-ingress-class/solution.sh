kubectl annotate ingressclass nginx ingressclass.kubernetes.io/is-default-class=true
kubectl delete -f /opt/course/nw-ingress-class/legacy.yaml
kubectl apply -f /opt/course/nw-ingress-class/legacy.yaml
sleep 5
