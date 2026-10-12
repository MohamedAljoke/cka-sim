listed() { tr -d ' \r' < "$COURSE/$1" 2>/dev/null | grep . | sort | xargs; }
crds_ok() { eq "alertrules.monitoring.cka.io dashboards.monitoring.cka.io probes.monitoring.cka.io" "$(listed crds.txt)"; }
dashboards_ok() { eq "ar-crd-list-team-a/cluster-overview ar-crd-list-team-a/node-health ar-crd-list-team-b/payments" "$(listed dashboards.txt)"; }

check 2 "crds.txt lists exactly the 3 CRDs of monitoring.cka.io" crds_ok
check 2 "dashboards.txt lists exactly the 3 Dashboards as namespace/name" dashboards_ok
