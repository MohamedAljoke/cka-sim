NS=nw-ingress-pathtype
HOST=docs.nw-ingress-pathtype.test
under() { eq "docs /docs/guide" "$(fetch $HOST /docs/guide)" && eq "docs /docs/a/b" "$(fetch $HOST /docs/a/b)"; }
exact_kept() { under && eq "docs /docs" "$(fetch $HOST /docs)" && ! fetch $HOST /documents | grep -q '^docs '; }

check 3 "/docs/guide and /docs/a/b reach docs" under
check 1 "/docs still works and /documents does not reach docs" exact_kept
check 1 "The path's pathType is Prefix" eq Prefix "$(kubectl -n $NS get ingress docs -o jsonpath='{.spec.rules[0].http.paths[0].pathType}')"
