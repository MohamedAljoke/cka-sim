SA=system:serviceaccount:ar-scope-dev:ci
can() { eq yes "$(kubectl auth can-i list pods -n "$1" --as=$SA 2>/dev/null)"; }
dev() { can ar-scope-dev; }
not_elsewhere() { dev && ! can ar-scope-prod && ! can default && ! can kube-system; }
via_rolebinding() { eq "ClusterRole/pod-reader" "$(kubectl -n ar-scope-dev get rolebindings -o jsonpath='{range .items[?(@.subjects[0].name=="ci")]}{.roleRef.kind}/{.roleRef.name}{end}')"; }
no_new_roles() { dev && [ -z "$(kubectl get roles -n ar-scope-dev -o name)" ]; }

check 2 "ci can list Pods in ar-scope-dev" dev
check 2 "ci can't list Pods in any other namespace" not_elsewhere
check 1 "A RoleBinding in ar-scope-dev grants ClusterRole pod-reader to ci" via_rolebinding
check 1 "No new Role was created" no_new_roles
