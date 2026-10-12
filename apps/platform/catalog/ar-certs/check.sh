crt=/etc/kubernetes/pki/apiserver.crt
day() { date -u -d "$1" +%F 2>/dev/null; }
ca_day() { day "$(openssl x509 -in /etc/kubernetes/pki/ca.crt -noout -enddate | cut -d= -f2)"; }
written_day() { day "$(tr -d , < "$COURSE/ca-expiry" | head -1)"; }
expiry_ok() { [ -n "$(written_day)" ] && eq "$(ca_day)" "$(written_day)"; }
started() { date -u -d "$(kubectl get namespace ar-certs -o jsonpath='{.metadata.creationTimestamp}')" +%s; }
issued() { date -u -d "$(openssl x509 -in $crt -noout -startdate | cut -d= -f2)" +%s; }
renewed() { [ "$(issued)" -ge "$(( $(started) - 300 ))" ] && openssl verify -CAfile /etc/kubernetes/pki/ca.crt $crt; }
served_serial() { openssl s_client -connect 127.0.0.1:6443 </dev/null 2>/dev/null | openssl x509 -noout -serial; }
served() { renewed && eq "$(openssl x509 -in $crt -noout -serial)" "$(served_serial)"; }
renewed >/dev/null 2>&1 && wait_for 60 served

check 2 "ca-expiry holds the cluster CA's expiry date" expiry_ok
check 2 "apiserver.crt was renewed and is signed by the cluster CA" renewed
check 2 "kube-apiserver serves the renewed certificate" served
