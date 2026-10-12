image=$(kubectl -n wl-image-digest get pods -l app=web -o jsonpath='{.items[0].status.containerStatuses[0].imageID}')
kubectl -n wl-image-digest set image deployment web web="$image" >/dev/null
kubectl -n wl-image-digest patch deployment web --type json -p '[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"IfNotPresent"}]' >/dev/null
echo "${image#*@}" > /opt/course/wl-image-digest/digest.txt
kubectl -n wl-image-digest rollout status deployment web --timeout=90s >/dev/null
