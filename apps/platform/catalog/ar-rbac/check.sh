SA=system:serviceaccount:ci:deployer
can() { eq yes "$(kubectl auth can-i "$1" "$2" -n ci --as=$SA 2>/dev/null)"; }
cannot() { ! can "$1" "$2"; }
can_deploy() { can get deployments.apps && can list deployments.apps && can create deployments.apps && can update deployments.apps; }
nothing_else() { cannot create pods && cannot get secrets && cannot delete services; }
least() { can_deploy && "$@"; }

check 1 "ServiceAccount ci/deployer exists" kubectl -n ci get serviceaccount deployer
check 1 "Role and RoleBinding deployer exist" kubectl -n ci get role/deployer rolebinding/deployer
check 2 "deployer can get, list, create and update deployments" can_deploy
check 1 "deployer cannot delete deployments" least cannot delete deployments.apps
check 1 "deployer has no access to other resources" least nothing_else
check 1 "can-delete.txt contains the can-i answer (no)" eq no "$(tr -d '[:space:]' < "$COURSE/can-delete.txt" 2>/dev/null)"
