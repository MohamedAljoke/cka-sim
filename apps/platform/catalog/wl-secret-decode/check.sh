token() { kubectl -n wl-secret-decode get secret api-token -o jsonpath='{.data.token}' | base64 -d; }
written() { tr -d '\r\n' < "$COURSE/token.txt" 2>/dev/null; }

check 2 "token.txt holds the decoded token" eq "$(token)" "$(written)"
