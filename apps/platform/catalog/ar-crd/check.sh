established() { eq True "$(kubectl get crd backups.cka.example.com -o jsonpath='{.status.conditions[?(@.type=="Established")].status}')"; }
cr() { kubectl -n ar-crd get backups.cka.example.com nightly -o jsonpath="$1"; }
cr_ok() { eq '0 2 * * *' "$(cr '{.spec.schedule}')" && eq 7 "$(cr '{.spec.retention}')"; }
fields_ok() { eq "retention schedule target" "$(grep -oE '\b(retention|schedule|target)\b' "$COURSE/fields.txt" | sort -u | xargs)"; }

check 2 "CRD backups.cka.example.com is installed" established
check 2 "Backup nightly exists with schedule 0 2 * * * and retention 7" cr_ok
check 1 "fields.txt lists schedule, retention and target" fields_ok
