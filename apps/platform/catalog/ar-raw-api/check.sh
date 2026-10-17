F="$COURSE/pods.json"
podlist() { eq PodList "$(jq -r .kind "$F" 2>/dev/null)"; }
names_ok() { podlist && eq "$(kubectl -n ar-raw-api get pods -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | sort | xargs)" "$(jq -r '.items[].metadata.name' "$F" | sort | xargs)"; }
replicas_ok() { eq 3 "$(tr -d ' \r\n' < "$COURSE/replicas.txt" 2>/dev/null)"; }

check 2 "pods.json is the PodList of ar-raw-api" names_ok
check 2 "replicas.txt holds 3" replicas_ok
