echo 'kubectl get pods -A --sort-by=.metadata.creationTimestamp' > /opt/course/tr-sort-by/by-age.sh
kubectl -n tr-sort-by get pods --sort-by='.status.containerStatuses[0].restartCount' -o name | tail -1 | cut -d/ -f2 > /opt/course/tr-sort-by/restarts.txt
