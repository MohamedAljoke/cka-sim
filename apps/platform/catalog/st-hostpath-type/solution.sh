# kubectl -n st-hostpath-type describe pod agent -> MountVolume.SetUp failed ... hostPath type check failed: /var/lib/agent-data is not a directory
f=/opt/course/st-hostpath-type/agent.yaml
sed -i 's/      type: Directory$/      type: DirectoryOrCreate/' "$f"
kubectl replace --force -f "$f"
kubectl -n st-hostpath-type wait --for=condition=Ready pod/agent --timeout=60s
