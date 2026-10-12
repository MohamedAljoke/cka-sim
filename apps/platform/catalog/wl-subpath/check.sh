NS=wl-subpath
D() { kubectl -n $NS get deployment app -o jsonpath="$1"; }
no_subpath() { [ -z "$(D '{.spec.template.spec.containers[0].volumeMounts[*].subPath}')" ] && [ -n "$(D '{.spec.template.spec.containers[0].volumeMounts[*].name}')" ]; }
ready() { no_subpath && eq "1 1" "$(D '{.status.updatedReplicas} {.status.readyReplicas}')" && [ -z "$(D '{.status.unavailableReplicas}')" ]; }
reads() { kubectl -n $NS exec deploy/app -- cat /etc/app/settings.conf 2>/dev/null; }
current() { ready && eq "level=debug" "$(reads)"; }
linked() { current && eq ..data/settings.conf "$(kubectl -n $NS exec deploy/app -- readlink /etc/app/settings.conf 2>/dev/null)"; }
no_subpath && wait_for 90 current

check 3 "app mounts the ConfigMap without subPath" no_subpath
check 1 "The running Pod reads level=debug at /etc/app/settings.conf" current
check 2 "settings.conf is the kubelet-managed link that follows ConfigMap updates" linked
