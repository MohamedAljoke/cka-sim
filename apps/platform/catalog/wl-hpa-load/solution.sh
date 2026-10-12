kubectl -n wl-hpa-load run load --image=busybox:1.36 -- sh -c 'while true; do wget -q -O- http://web >/dev/null; done' >/dev/null
echo 420 > /opt/course/wl-hpa-load/scale-down.txt
scaled() { [ "$(kubectl -n wl-hpa-load get deployment web -o jsonpath='{.spec.replicas}')" -ge 2 ]; }
wait_for 170 scaled
