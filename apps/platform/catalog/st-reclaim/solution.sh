pv=$(kubectl -n st-reclaim get pvc orders -o jsonpath='{.spec.volumeName}')
printf 'pv: %s\npolicy: %s\n' "$pv" "$(kubectl get pv "$pv" -o jsonpath='{.spec.persistentVolumeReclaimPolicy}')" > /opt/course/st-reclaim/pv.txt
kubectl patch pv "$pv" -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'
