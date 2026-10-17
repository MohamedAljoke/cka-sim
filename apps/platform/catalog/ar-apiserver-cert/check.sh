crt=/etc/kubernetes/pki/apiserver.crt
answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/cert.txt" 2>/dev/null | head -1 | tr -d '\r'; }
issuer() { openssl x509 -in $crt -noout -issuer | sed -n 's/.*CN *= *\([^,]*\).*/\1/p'; }
day() { date -u -d "$1" +%F 2>/dev/null; }
expires_ok() { local w; w=$(day "$(answer expires | tr -d ,)"); [ -n "$w" ] && eq "$(day "$(openssl x509 -in $crt -noout -enddate | cut -d= -f2)")" "$w"; }
names() { openssl x509 -in $crt -noout -ext subjectAltName | grep -o 'DNS:[^,]*' | cut -d: -f2 | sort | paste -sd,; }
covered() { names | tr ',' '\n' | grep -qx k8s.cka.internal && echo yes || echo no; }

check 1 "issuer is the CA's CN" eq "$(issuer)" "$(answer issuer | xargs)"
check 1 "expires is the certificate's expiry date" expires_ok
check 2 "dns-names lists every DNS name in the SANs" eq "$(names)" "$(answer dns-names | tr -d ' ' | tr ',' '\n' | sort | paste -sd,)"
check 1 "k8s.cka.internal says whether that name is covered" eq "$(covered)" "$(answer k8s.cka.internal | xargs)"
