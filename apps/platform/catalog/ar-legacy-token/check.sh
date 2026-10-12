NS=ar-legacy-token
S() { kubectl -n $NS get secret deployer-token -o jsonpath="$1" 2>/dev/null; }
typed() { eq "kubernetes.io/service-account-token deployer" "$(S '{.type} {.metadata.annotations.kubernetes\.io/service-account\.name}')"; }
filled() { typed && [ -n "$(S '{.data.token}')" ]; }
token() { tr -d ' \r\n' < "$COURSE/token" 2>/dev/null; }
written() { filled && eq "$(S '{.data.token}' | base64 -d)" "$(token)"; }
works() { written && eq system:serviceaccount:$NS:deployer "$(kubectl --token="$(token)" --server=https://127.0.0.1:6443 --certificate-authority=/etc/kubernetes/pki/ca.crt --kubeconfig=/dev/null auth whoami -o jsonpath='{.status.userInfo.username}' 2>/dev/null)"; }
typed && wait_for 20 filled

check 2 "Secret deployer-token is a service-account-token Secret for deployer" typed
check 1 "Kubernetes filled in its token" filled
check 2 "token holds the decoded token, which logs in as deployer" works
