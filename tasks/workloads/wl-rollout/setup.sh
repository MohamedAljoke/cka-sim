fresh_ns release
on_host "rm -rf $(course_dir) && mkdir -p $(course_dir)"
k -n release create deployment shop --image=nginx:1.26-alpine --replicas=3 >/dev/null
k -n release rollout status deploy/shop --timeout=180s >/dev/null
k -n release set image deploy/shop nginx=nginx:1.27-alpine >/dev/null
k -n release rollout status deploy/shop --timeout=180s >/dev/null
k -n release set image deploy/shop nginx=nginx:1.27-alpine-broken >/dev/null
