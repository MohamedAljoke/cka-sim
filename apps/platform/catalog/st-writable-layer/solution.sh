f=/opt/course/st-writable-layer/scratch.yaml
cat >> "$f" <<'YAML'
    volumeMounts:
    - name: cache
      mountPath: /cache
  volumes:
  - name: cache
    emptyDir: {}
YAML
kubectl replace --force -f "$f"
kubectl -n st-writable-layer wait --for=condition=Ready pod/scratch --timeout=60s
sleep 2
kubectl -n st-writable-layer exec scratch -- kill 1
until [ "$(kubectl -n st-writable-layer get pod scratch -o jsonpath='{.status.containerStatuses[0].restartCount}')" -ge 1 ] && \
      [ "$(kubectl -n st-writable-layer get pod scratch -o jsonpath='{.status.containerStatuses[0].ready}')" = true ]; do sleep 2; done
kubectl -n st-writable-layer exec scratch -- cat /cache/marker
