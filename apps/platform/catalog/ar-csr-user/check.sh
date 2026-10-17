J() { kubectl --context jane --request-timeout=5s "$@"; }
approved() { kubectl get csr jane -o jsonpath='{.status.conditions[?(@.type=="Approved")].status} {.spec.signerName}' | grep -qx 'True kubernetes.io/kube-apiserver-client'; }
cert_ok() { openssl verify -CAfile /etc/kubernetes/pki/ca.crt "$COURSE/jane.crt" && openssl x509 -in "$COURSE/jane.crt" -noout -subject | grep -q 'CN *= *jane'; }
pair() { cert_ok && eq "$(openssl x509 -in "$COURSE/jane.crt" -noout -pubkey)" "$(openssl pkey -in "$COURSE/jane.key" -pubout)"; }
whoami() { J auth whoami -o jsonpath='{.status.userInfo.username}'; }
context_ok() { eq jane "$(whoami)" && eq ar-csr-user "$(kubectl config view -o jsonpath='{.contexts[?(@.name=="jane")].context.namespace}')"; }
reads() { context_ok && J get configmaps app-settings; }
read_only() { reads && ! J create configmap probe --dry-run=server && ! J -n default get pods && ! J get secrets; }

check 2 "CSR jane is approved for signer kube-apiserver-client" approved
check 2 "jane.crt is issued by the cluster CA for CN=jane and matches jane.key" pair
check 1 "Context jane logs in as jane, defaulting to Namespace ar-csr-user" context_ok
check 2 "jane can read in ar-csr-user" reads
check 1 "jane can't write, read Secrets, or read other namespaces" read_only
