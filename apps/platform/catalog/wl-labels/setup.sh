fresh_ns wl-labels
fresh_course
pod() { kubectl -n wl-labels run "$1" --image=nginx:1.27 ${2:+--labels="$2"} >/dev/null; }
pod web-prod app=web,tier=frontend,env=prod
pod web-dev app=web,tier=frontend,env=dev
pod api app=api,tier=backend,env=prod
pod db app=db,tier=backend,env=prod
pod cache app=cache,tier=backend,env=dev
pod debug
kubectl -n wl-labels wait pod --all --for=condition=Ready --timeout=120s >/dev/null
