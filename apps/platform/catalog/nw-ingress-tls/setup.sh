fresh_ns nw-ingress-tls
fresh_course
openssl req -x509 -newkey rsa:2048 -nodes -days 30 -subj /CN=secure.nw-ingress-tls.test \
  -addext subjectAltName=DNS:secure.nw-ingress-tls.test \
  -keyout "$COURSE/tls.key" -out "$COURSE/tls.crt" 2>/dev/null
web_app nw-ingress-tls secure
kubectl -n nw-ingress-tls create ingress secure --class=nginx --rule='secure.nw-ingress-tls.test/*=secure:80' >/dev/null
