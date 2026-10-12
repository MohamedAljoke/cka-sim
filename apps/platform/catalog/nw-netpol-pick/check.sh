NS=nw-netpol-pick
DB=$(pod_ip $NS db)
backend_only() { connects $NS backend "$DB" 5432 && ! connects $NS frontend "$DB" 5432; }
one_port() { ! connects $NS backend "$DB" 9000; }
one_ns() { ! connects nw-netpol-pick-other backend "$DB" 5432; }
policies() { kubectl -n $NS get networkpolicy -o jsonpath='{.items[*].metadata.name}'; }

check 2 "backend reaches db:5432, frontend is blocked" backend_only
check 1 "db:9000 is blocked even for backend" one_port
check 1 "A backend Pod in nw-netpol-pick-other is blocked" one_ns
check 1 "Only db-c is applied" eq db-c "$(policies)"
