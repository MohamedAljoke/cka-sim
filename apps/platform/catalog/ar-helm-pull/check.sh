chart=$COURSE/web
meta() { helm -n ar-helm-local get metadata local 2>/dev/null | awk -F': ' -v key="$1" '$1 == key { print $2 }'; }
unpacked() { grep -qx 'version: 1.1.0' "$chart/Chart.yaml" && [ -d "$chart/templates" ]; }
four() { grep -qx 'replicaCount: 4' "$chart/values.yaml"; }
installed() { eq "deployed 1.1.0" "$(meta STATUS) $(meta VERSION)"; }
no_overrides() { helm -n ar-helm-local get values local -o json 2>/dev/null | grep -qx 'null\|{}'; }
running() { eq 4 "$(kubectl -n ar-helm-local get deployment local-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)"; }

check 1 "web 1.1.0 is unpacked in /opt/course/ar-helm-pull/web" unpacked
check 1 "the chart's values.yaml sets replicaCount 4" four
check 1 "release local of web 1.1.0 is deployed in ar-helm-local" installed
check 1 "release local overrides no value" no_overrides
check 1 "Deployment local-web runs 4 Pods" wait_for 60 running
