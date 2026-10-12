can() { eq yes "$(kubectl auth can-i "$@" --as=contractor 2>/dev/null)"; }
listed() { tr -d ' \r' < "$COURSE/bindings.txt" 2>/dev/null | sed 's|.*/||' | grep . | sort | xargs; }
found() { eq "contractor-debug contractor-view deploy-helpers oncall" "$(listed)"; }
reads() { can list pods -n ar-audit-app && can get configmaps -n ar-audit-app && ! can list nodes; }
only_reads() { reads && ! can delete pods -n ar-audit-app && ! can create deployments -n ar-audit-app && ! can get secrets -n ar-audit-app && ! can list pods -n ar-audit-ops && ! can list nodes; }
others_kept() { only_reads && eq yes "$(kubectl auth can-i create deployments -n ar-audit-app --as=system:serviceaccount:ar-audit-app:default)" && eq yes "$(kubectl auth can-i delete pods -n ar-audit-ops --as=someone --as-group=oncall)"; }

check 2 "bindings.txt names all 4 bindings that include contractor" found
check 1 "contractor lost cluster-wide rights but can still read in ar-audit-app" reads
check 2 "contractor can do nothing more" only_reads
check 1 "The other subjects of shared bindings kept their rights" others_kept
