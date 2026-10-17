NS=nw-connect
reply() { eq "backend backend says hello" "$(tr -d '\r' < "$COURSE/$1" 2>/dev/null | grep . | head -1)"; }
by_ip() { reply by-ip.txt; }
by_name() { reply by-name.txt; }
only_backend() { [ "$(kubectl -n $NS get pods -o name)" = "pod/backend" ]; }
tidy() { by_name && wait_for 30 only_backend; }

check 2 "by-ip.txt holds the reply of Pod backend" by_ip
check 2 "by-name.txt holds the reply through Service backend" by_name
check 1 "The temporary Pod is gone" tidy
