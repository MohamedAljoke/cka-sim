ver() { kubectl get crd gadgets.cka.example.com -o json | jq -c "[.spec.versions[] | {name, served, storage}] | sort_by(.name)"; }
versions_ok() { eq '[{"name":"v1","served":true,"storage":true},{"name":"v1alpha1","served":true,"storage":false}]' "$(ver)"; }
beacon() { versions_ok && eq "cka.example.com/v1 blue" "$(kubectl -n ar-crd-versions get gadgets.v1.cka.example.com beacon -o jsonpath='{.apiVersion} {.spec.color}' 2>/dev/null)"; }
both_served() { beacon && eq blue "$(kubectl -n ar-crd-versions get gadgets.v1alpha1.cka.example.com beacon -o jsonpath='{.spec.color}' 2>/dev/null)"; }
stored() { kubectl get crd gadgets.cka.example.com -o jsonpath='{.status.storedVersions}' | jq -r 'sort | join(",")'; }
written() { tr -d ' \r\n' < "$COURSE/stored.txt" 2>/dev/null | tr ',' '\n' | sort | paste -sd,; }
stored_ok() { versions_ok && eq "v1,v1alpha1" "$(stored)" && eq "$(stored)" "$(written)"; }

check 2 "v1 is served and stored, v1alpha1 served but not stored" versions_ok
check 2 "Gadget beacon exists through v1 and reads the same through v1alpha1" both_served
check 1 "stored.txt lists the stored versions v1 and v1alpha1" stored_ok
