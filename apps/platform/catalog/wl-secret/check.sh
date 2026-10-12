NS=wl-secret
val() { kubectl -n $NS get secret db-creds -o jsonpath="{.data.$1}" | base64 -d; }
secret_ok() { eq admin "$(val username)" && eq 'S3cr3t!' "$(val password)"; }
running() { eq True "$(kubectl -n $NS get pod db-client -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')"; }
in_pod() { kubectl -n $NS exec db-client -- "$@"; }
env_ok() { running && eq admin "$(in_pod printenv username)" && eq 'S3cr3t!' "$(in_pod printenv password)" && eq db-creds "$(kubectl -n $NS get pod db-client -o jsonpath='{.spec.containers[0].envFrom[0].secretRef.name}')"; }
mount() { kubectl -n $NS get pod db-client -o jsonpath='{range .spec.containers[0].volumeMounts[?(@.mountPath=="/etc/db-creds")]}{.readOnly}{end}'; }
files_ok() { running && eq 'S3cr3t!' "$(in_pod cat /etc/db-creds/password)" && eq true "$(mount)"; }
secret_ok && wait_for 60 running

check 2 "Secret db-creds holds username=admin and password=S3cr3t!" secret_ok
check 2 "db-client gets the keys as env vars through envFrom" env_ok
check 2 "db-client mounts db-creds read-only at /etc/db-creds" files_ok
