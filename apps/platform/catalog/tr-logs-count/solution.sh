for p in $(kubectl -n tr-logs-count get pods -l app=worker -o name); do
  echo "${p#pod/}: $(kubectl -n tr-logs-count logs "$p" | grep -c ERROR)"
done > /opt/course/tr-logs-count/errors.txt
