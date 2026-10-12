SA=system:serviceaccount:ar-resourcenames:toggler
can() { eq yes "$(kubectl auth can-i "$@" -n ar-resourcenames --as=$SA 2>/dev/null)"; }
on_flags() { can get configmap/feature-flags && can update configmap/feature-flags; }
only_flags() { on_flags && ! can get configmap/db-config && ! can update configmap/db-config && ! can list configmaps && ! can delete configmap/feature-flags && ! can get secrets; }
bound() { eq "Role/flags-editor" "$(kubectl -n ar-resourcenames get rolebinding toggler-flags -o jsonpath='{.roleRef.kind}/{.roleRef.name}' 2>/dev/null)"; }

check 2 "toggler can get and update feature-flags" on_flags
check 2 "toggler can't touch other ConfigMaps, list them, or delete feature-flags" only_flags
check 1 "RoleBinding toggler-flags uses Role flags-editor" bound
