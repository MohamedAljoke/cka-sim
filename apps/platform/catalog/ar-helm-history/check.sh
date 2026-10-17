version() { tr -d '[:space:]' < "$COURSE/version.txt" 2>/dev/null | sed 's/^web-//'; }
values_of_2() { grep -qE '^replicaCount: 2$' "$COURSE/values.yaml" && grep -qE '^\s+type: NodePort$' "$COURSE/values.yaml" && ! grep -qE '^replicaCount: 1$' "$COURSE/values.yaml"; }

check 2 "version.txt says 1.1.0" eq 1.1.0 "$(version)"
check 3 "values.yaml holds revision 2's values: replicaCount 2 and service type NodePort" values_of_2
