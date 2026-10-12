NS=ar-kustomize-gen
D() { kubectl -n $NS get deployment app -o jsonpath="$1"; }
cm() { D '{.spec.template.spec.volumes[0].configMap.name}'; }
secret() { D '{.spec.template.spec.containers[0].env[0].valueFrom.secretKeyRef.name}'; }
generated() { [[ "$(cm)" == app-config-* ]] && [[ "$(secret)" == app-secret-* ]]; }
password() { kubectl -n $NS get secret "$(secret)" -o jsonpath='{.data.password}' | base64 -d; }
secret_ok() { generated && eq s3cr3t "$(password)"; }
ready() { eq 1 "$(D '{.status.readyReplicas}')" && eq 1 "$(D '{.status.updatedReplicas}')" && eq "$(D '{.metadata.generation}')" "$(D '{.status.observedGeneration}')"; }
sees_debug() { generated && ready && kubectl -n $NS exec deploy/app -- cat /etc/app/app.properties | grep -qx 'mode=debug'; }
rolled() { sees_debug && [ "$(kubectl -n $NS get replicasets -l app=app -o name | wc -l)" -ge 2 ]; }
generated && wait_for 90 sees_debug

check 2 "app uses a generated ConfigMap app-config-<hash> and Secret app-secret-<hash>" generated
check 1 "The generated Secret holds password=s3cr3t" secret_ok
check 2 "The running Pod of app sees mode=debug in /etc/app/app.properties" sees_debug
check 1 "The change rolled out a new ReplicaSet" rolled
