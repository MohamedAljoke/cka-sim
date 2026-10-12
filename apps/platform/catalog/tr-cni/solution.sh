# kubectl -n tr-cni describe pod -l app=web -> FailedCreatePodSandBox: failed to find plugin "ptpp" in path [/opt/cni/bin]
# ls /etc/cni/net.d/ /opt/cni/bin/          -> the conflist asks for "ptpp", the binary is "ptp"
sed -i 's/"type": "ptpp"/"type": "ptp"/' /etc/cni/net.d/10-kindnet.conflist
