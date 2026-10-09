policy() { k -n payments get networkpolicy api-allow; }
reach() { k -n "$1" exec "$2" -- wget -qO- -T 3 http://api.payments >/dev/null; }
# Allowed traffic only scores with the policy in place: without one, everything is "reachable".
allowed() { policy && reach "$1" "$2"; }
blocked() { ! reach "$1" "$2"; }
check 1 "NetworkPolicy payments/api-allow exists" policy
check 2 "payments/frontend can reach api" allowed payments frontend
check 2 "monitoring/prober can reach api" allowed monitoring prober
check 2 "payments/batch is blocked" blocked payments batch
