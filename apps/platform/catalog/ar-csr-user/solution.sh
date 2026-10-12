cd /opt/course/ar-csr-user
openssl genrsa -out jane.key 2048
openssl req -new -key jane.key -subj "/CN=jane" -out jane.csr
cat <<YAML | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: jane
spec:
  request: $(base64 -w0 jane.csr)
  signerName: kubernetes.io/kube-apiserver-client
  usages: [client auth]
YAML
kubectl certificate approve jane
until [ -n "$(kubectl get csr jane -o jsonpath='{.status.certificate}')" ]; do sleep 1; done
kubectl get csr jane -o jsonpath='{.status.certificate}' | base64 -d > jane.crt
kubectl config set-credentials jane --client-key=jane.key --client-certificate=jane.crt --embed-certs
kubectl config set-context jane --cluster=cka-sim --user=jane --namespace=ar-csr-user
kubectl -n ar-csr-user create rolebinding jane-view --clusterrole=view --user=jane
kubectl --context jane get configmaps
