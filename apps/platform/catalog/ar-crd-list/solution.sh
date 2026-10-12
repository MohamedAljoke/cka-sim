kubectl get crd -o name | cut -d/ -f2 | grep '\.monitoring\.cka\.io$' > /opt/course/ar-crd-list/crds.txt
kubectl get dashboards.monitoring.cka.io -A -o custom-columns=NS:.metadata.namespace,NAME:.metadata.name --no-headers |
  awk '{print $1"/"$2}' > /opt/course/ar-crd-list/dashboards.txt
