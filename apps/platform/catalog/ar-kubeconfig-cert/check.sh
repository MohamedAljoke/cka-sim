answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/identity.txt" 2>/dev/null | head -1 | tr -d '\r'; }
cert() { kubectl --kubeconfig "$COURSE/auditor.kubeconfig" config view --raw -o jsonpath='{.users[0].user.client-certificate-data}' | base64 -d; }
subject() { cert | openssl x509 -noout -subject -nameopt RFC2253; }
cn() { subject | grep -o 'CN=[^,]*' | cut -d= -f2; }
groups() { subject | grep -o 'O=[^,]*' | cut -d= -f2 | sort | paste -sd,; }
day() { date -u -d "$1" +%F 2>/dev/null; }
expires_ok() { local w; w=$(day "$(answer expires | tr -d ,)"); [ -n "$w" ] && eq "$(day "$(cert | openssl x509 -noout -enddate | cut -d= -f2)")" "$w"; }

check 1 "user is the certificate's CN" eq "$(cn)" "$(answer user | xargs)"
check 2 "groups lists every O" eq "$(groups)" "$(answer groups | tr -d ' ' | tr ',' '\n' | sort | paste -sd,)"
check 1 "expires is its expiry date" expires_ok
