# journalctl -u kubelet | grep node-exporter -> could not process manifest file ... yaml: line 10: mapping values are not allowed
sed -i 's/^     command:/    command:/' /etc/kubernetes/manifests/node-exporter.yaml
