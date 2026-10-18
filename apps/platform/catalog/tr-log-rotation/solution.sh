printf 'containerLogMaxSize: 5Mi\ncontainerLogMaxFiles: 3\n' >> /var/lib/kubelet/config.yaml
systemctl restart kubelet
