NS=nw-netpol-ports
SERVER=$(pod_ip $NS server)
range() { connects $NS client "$SERVER" 8000 && connects $NS client "$SERVER" 8005 && connects $NS client "$SERVER" 8010 && ! connects $NS client "$SERVER" 8011; }
named() { connects $NS client "$SERVER" 9100 && ! connects $NS client "$SERVER" 8011; }
ports() { kubectl -n $NS get networkpolicy server-ports -o jsonpath='{range .spec.ingress[*].ports[*]}{.port}-{.endPort} {end}' | xargs -n1 | sort | xargs; }

check 2 "client reaches 8000, 8005 and 8010, but not 8011" range
check 1 "client reaches metrics (9100)" named
check 2 "The ports are 8000 with endPort 8010, and metrics by name" eq "8000-8010 metrics-" "$(ports)"
