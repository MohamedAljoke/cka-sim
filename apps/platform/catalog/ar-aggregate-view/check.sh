cr() { kubectl get clusterrole widget-viewer -o jsonpath="$1" 2>/dev/null; }
labelled() { eq true "$(cr '{.metadata.labels.rbac\.authorization\.k8s\.io/aggregate-to-view}')"; }
in_view() { kubectl get clusterrole view -o json | jq -e '.rules[] | select(.resources | index("widgets")) | select(.apiGroups | index("cka.example.com"))' >/dev/null; }
reads() { labelled && in_view && eq yes "$(kubectl auth can-i list widgets.cka.example.com -n ar-aggregate --as=reader)"; }
only_reads() { reads && eq no "$(kubectl auth can-i delete widgets.cka.example.com -n ar-aggregate --as=reader)"; }
view_untouched() { reads && eq true "$(kubectl get clusterrole view -o jsonpath='{.metadata.annotations.rbac\.authorization\.kubernetes\.io/autoupdate}')"; }
labelled && wait_for 20 in_view

check 2 "widget-viewer carries the aggregate-to-view label" labelled
check 2 "reader can list widgets in ar-aggregate through view" reads
check 1 "reader still can't change widgets" only_reads
