names() { kubectl -n wl-labels get pods "$@" -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | sort | xargs; }
listed() { sort "$COURSE/backend-prod.txt" 2>/dev/null | xargs; }
label() { kubectl -n wl-labels get pod "$1" -o jsonpath="{.metadata.labels.$2}"; }
env_removed() { [ -z "$(label web-dev env)" ] && eq web "$(label web-dev app)"; }
owner() { kubectl -n wl-labels get pod api -o jsonpath='{.metadata.annotations.owner}'; }

check 2 "backend-prod.txt lists api and db" eq "api db" "$(listed)"
check 2 "Exactly web-dev and web-prod carry team=payments" eq "web-dev web-prod" "$(names -l team=payments)"
check 1 "web-dev has no env label and keeps app=web" env_removed
check 1 "api is annotated owner=platform" eq platform "$(owner)"
