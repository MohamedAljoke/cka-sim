secret() { kubectl -n tr-private-registry get secret registry-creds -o jsonpath="$1" 2>/dev/null; }
config() { secret '{.data.\.dockerconfigjson}' | base64 -d 2>/dev/null; }
creds() { eq kubernetes.io/dockerconfigjson "$(secret '{.type}')" && config | grep -q '"cka-sim-registry:5000"' && config | grep -q '"username":"candidate"'; }
on_sa() { kubectl -n tr-private-registry get serviceaccount default -o jsonpath='{.imagePullSecrets[*].name}' | grep -qw registry-creds; }
template_unchanged() { [ -z "$(kubectl -n tr-private-registry get deployment web -o jsonpath='{.spec.template.spec.imagePullSecrets}')" ]; }
running() { eq 2 "$(kubectl -n tr-private-registry get deployment web -o jsonpath='{.status.readyReplicas}')"; }

check 2 "Secret registry-creds holds the login for cka-sim-registry:5000" creds
check 2 "ServiceAccount default pulls images with registry-creds" on_sa
check 3 "both Pods of web run, with no imagePullSecrets in web's Pod template" eval 'template_unchanged && wait_for 90 running'
