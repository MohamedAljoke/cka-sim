NS=nw-ingress-rewrite
HOST=app.nw-ingress-rewrite.test
annotation() { kubectl -n $NS get ingress app -o jsonpath='{.metadata.annotations.nginx\.ingress\.kubernetes\.io/rewrite-target}'; }

check 3 "/app on $HOST reaches web as /" eq "web /" "$(fetch $HOST /app)"
check 2 "Ingress app has the rewrite-target annotation set to /" eq / "$(annotation)"
