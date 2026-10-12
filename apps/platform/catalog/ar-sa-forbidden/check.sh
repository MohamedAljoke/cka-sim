SA=system:serviceaccount:ar-sa:reporter
can() { eq yes "$(kubectl auth can-i "$@" --as=$SA 2>/dev/null)"; }
lists() { can list pods -n ar-sa; }
nothing_more() { lists && ! can delete pods -n ar-sa && ! can get secrets -n ar-sa && ! can list pods -n default && ! can list pods -A; }
untouched() { lists && eq 1 "$(kubectl -n ar-sa get deployment reporter -o jsonpath='{.metadata.generation}')"; }

check 3 "reporter can list Pods in ar-sa" lists
check 2 "reporter can't delete Pods, read Secrets or list Pods elsewhere" nothing_more
check 1 "Deployment reporter was not modified" untouched
