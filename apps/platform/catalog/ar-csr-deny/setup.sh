kubectl delete csr dev-anna ci-runner --ignore-not-found >/dev/null
csr() {
  local key req
  key=$(mktemp)
  req=$(mktemp)
  openssl genrsa -out "$key" 2048 2>/dev/null
  openssl req -new -key "$key" -subj "$2" -out "$req" 2>/dev/null
  kubectl apply -f - >/dev/null <<YAML
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: $1
spec:
  request: $(base64 -w0 "$req")
  signerName: kubernetes.io/kube-apiserver-client
  usages: [client auth]
YAML
  rm -f "$key" "$req"
}
csr dev-anna "/CN=anna/O=developers"
csr ci-runner "/CN=ci-runner/O=kubeadm:cluster-admins"
