NS=wl-pdb
pdb() { kubectl -n $NS get pdb api-pdb -o jsonpath="$1"; }
spec_ok() { eq 2 "$(pdb '{.spec.minAvailable}')" || eq 1 "$(pdb '{.spec.maxUnavailable}')"; }
selects() { eq 3 "$(pdb '{.status.expectedPods}')"; }
allowed() { eq 1 "$(pdb '{.status.disruptionsAllowed}')"; }
guards() { spec_ok && selects && allowed; }
spec_ok && wait_for 30 guards

check 2 "api-pdb keeps at least 2 Pods available" spec_ok
check 2 "api-pdb selects the 3 Pods of api" selects
check 1 "api-pdb allows exactly 1 disruption right now" guards
