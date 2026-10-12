Kubernetes has no user objects. A user is whoever presents a client certificate signed by the
cluster CA: the certificate's **CN** is the user name and each **O** is a group. Through the
**Certificates API** you get such a certificate without touching the CA key:

```sh
openssl genrsa -out jane.key 2048
openssl req -new -key jane.key -subj "/CN=jane" -out jane.csr
# CertificateSigningRequest: spec.request = base64 of jane.csr,
#   signerName: kubernetes.io/kube-apiserver-client, usages: [client auth]
kubectl certificate approve jane
kubectl get csr jane -o jsonpath='{.status.certificate}' | base64 -d > jane.crt
kubectl config set-credentials jane --client-key=jane.key --client-certificate=jane.crt --embed-certs
kubectl config set-context jane --cluster=cka-sim --user=jane --namespace=ar-csr-user
kubectl -n ar-csr-user create rolebinding jane-view --clusterrole=view --user=jane
kubectl --context jane get all
```

The traps:
- `spec.request` must be the base64 of the whole PEM file, on one line (`base64 -w0`).
- `usages` must include `client auth`, or the API server won't accept the certificate.
- An approved CSR is signed a moment later. Read `.status.certificate` after it appears.
- `view` is a ClusterRole, but the **RoleBinding** limits it to one namespace. `view`
  deliberately can't read Secrets.

The kubernetes.io docs page "Certificates and Certificate Signing Requests" has a copy-ready CSR
YAML, worth bookmarking for the exam.
