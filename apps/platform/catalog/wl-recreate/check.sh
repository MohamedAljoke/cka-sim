NS=wl-recreate
D() { kubectl -n $NS get deployment db-migrator -o jsonpath="$1"; }
recreate() { eq Recreate "$(D '{.spec.strategy.type}')"; }
updated() { eq busybox:1.36 "$(D '{.spec.template.spec.containers[0].image}')" && eq 2 "$(D '{.status.updatedReplicas}')" && eq 2 "$(D '{.status.readyReplicas}')"; }
command_kept() { updated && eq '["sh","-c","sleep 1d"]' "$(D '{.spec.template.spec.containers[0].command}')"; }
recreate && wait_for 90 updated

check 2 "db-migrator uses the Recreate strategy" recreate
check 2 "db-migrator runs busybox:1.36 with 2 Ready Pods" updated
check 1 "The command is unchanged" command_kept
