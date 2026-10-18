# journalctl -u kubelet | grep -i unauthorized -> the API server rejects the kubelet's credentials
# grep client- /etc/kubernetes/kubelet.conf     -> /var/lib/kubelet/pki/kubelet-client-current.pem
# openssl x509 -in <pem> -noout -issuer         -> the newest cert is self-signed, the older one is issued by "kubernetes"
cd /var/lib/kubelet/pki
for f in kubelet-client-2*.pem; do
  openssl verify -CAfile /etc/kubernetes/pki/ca.crt "$f" >/dev/null 2>&1 && good=$f
done
ln -sfn "/var/lib/kubelet/pki/$good" kubelet-client-current.pem
systemctl restart kubelet
