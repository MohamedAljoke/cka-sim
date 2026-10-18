uid=$(kubectl -n st-node-paths get pod cache -o jsonpath='{.metadata.uid}')
id=$(crictl ps -q --name '^cache$' --namespace st-node-paths | head -1)
{
  echo "emptydir: /var/lib/kubelet/pods/$uid/volumes/kubernetes.io~empty-dir/scratch"
  echo "pid: $(crictl inspect "$id" | jq -r .info.pid)"
} > /opt/course/st-node-paths/paths.txt
