can() { eq yes "$(kubectl auth can-i "$@" --as=audit-user --as-group=auditors 2>/dev/null)"; }
cannot() { eq no "$(kubectl auth can-i "$@" --as=audit-user --as-group=auditors 2>/dev/null)"; }
reads() { can get nodes && can list nodes && can get persistentvolumes && can list persistentvolumes; }
only_reads() { reads && cannot delete nodes && cannot watch nodes && cannot create persistentvolumes && cannot list pods -A && cannot list secrets -A; }
bound() { eq "ClusterRole/node-viewer Group/auditors" "$(kubectl get clusterrolebinding node-viewer -o jsonpath='{.roleRef.kind}/{.roleRef.name} {.subjects[0].kind}/{.subjects[0].name}')"; }

check 2 "Group auditors can get and list Nodes and PersistentVolumes" reads
check 2 "Group auditors can do nothing else" only_reads
check 1 "ClusterRoleBinding node-viewer binds ClusterRole node-viewer to group auditors" bound
