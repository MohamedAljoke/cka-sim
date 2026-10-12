node=$(hostname)
N() { kubectl get node "$node" -o jsonpath="$1"; }
ki() { local v=$1; case $v in *Ki) echo "${v%Ki}" ;; *Mi) echo $(( ${v%Mi} * 1024 )) ;; *Gi) echo $(( ${v%Gi} * 1048576 )) ;; *) echo $(( v / 1024 )) ;; esac; }
milli() { local v=$1; case $v in *m) echo "${v%m}" ;; *) echo $(( v * 1000 )) ;; esac; }
configz() { kubectl get --raw "/api/v1/nodes/$node/proxy/configz" | jq -r ".kubeletconfig.$1"; }
ready() { eq True "$(N '{.status.conditions[?(@.type=="Ready")].status}')"; }
cpu_ok() { eq $(( $(milli "$(N '{.status.capacity.cpu}')") - 500 )) "$(milli "$(N '{.status.allocatable.cpu}')")"; }
mem_ok() { eq $(( $(ki "$(N '{.status.capacity.memory}')") - 524288 - 307200 )) "$(ki "$(N '{.status.allocatable.memory}')")"; }
reserved() { ready && cpu_ok && eq "500m 512Mi" "$(configz 'systemReserved.cpu') $(configz 'systemReserved.memory')"; }
eviction() { ready && eq 300Mi "$(configz 'evictionHard["memory.available"]')" && eq "0% 0%" "$(configz 'evictionHard["nodefs.available"]') $(configz 'evictionHard["imagefs.available"]')"; }
allocatable() { ready && cpu_ok && mem_ok; }
grep -q '^systemReserved:' /var/lib/kubelet/config.yaml && wait_for 90 allocatable

check 2 "The running kubelet reserves 500m CPU and 512Mi memory for the system" reserved
check 2 "Hard eviction at memory.available<300Mi, other thresholds kept" eviction
check 2 "Node cka-sim-worker2 is Ready and reports the reduced Allocatable" allocatable
