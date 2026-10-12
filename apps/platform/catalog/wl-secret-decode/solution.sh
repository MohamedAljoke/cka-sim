kubectl -n wl-secret-decode get secret api-token -o jsonpath='{.data.token}' | base64 -d > /opt/course/wl-secret-decode/token.txt
