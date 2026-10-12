f=/opt/course/tr-native-sidecar/report.yaml
sed -i 's/^spec:$/spec:\n  initContainers:\n  - name: log-shipper\n    image: busybox:1.36\n    restartPolicy: Always\n    command: [sh, -c, "tail -n+1 -F \/var\/log\/report\/report.log"]\n    volumeMounts:\n    - name: logs\n      mountPath: \/var\/log\/report/' "$f"
kubectl replace --force -f "$f"
kubectl -n tr-native-sidecar wait --for=condition=Ready pod/report --timeout=90s
