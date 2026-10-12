for p in payments search nightly; do
  echo "$p: $(kubectl -n tr-qos get pod $p -o jsonpath='{.status.qosClass}')"
done > /opt/course/tr-qos/classes.txt
f=/opt/course/tr-qos/payments.yaml
sed -i 's/cpu: 100m/cpu: 200m/; s/memory: 64Mi/memory: 128Mi/' "$f"
printf '      limits:\n        cpu: 200m\n        memory: 128Mi\n' >> "$f"
kubectl replace --force -f "$f"
kubectl -n tr-qos wait --for=condition=Ready pod/payments --timeout=60s
