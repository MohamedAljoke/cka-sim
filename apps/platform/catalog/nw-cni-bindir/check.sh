NS=nw-cni-bindir
cfg=/etc/containerd/config.toml
bin_dir() { sed -n 's/^ *bin_dirs\? *= *\[\?"\([^"]*\)".*/\1/p' "$cfg" | head -1; }
dir_ok() { [ -z "$(bin_dir)" ] || eq /opt/cni/bin "$(bin_dir)"; }
running() { eq 1 "$(kubectl -n $NS get deployment web -o jsonpath='{.status.readyReplicas}')"; }
untouched() { running && eq 1 "$(kubectl -n $NS get deployment web -o jsonpath='{.metadata.generation}')" && [ ! -e /opt/cni/plugins ]; }
dir_ok && wait_for 120 running

check 2 "containerd looks for CNI plugins in /opt/cni/bin" dir_ok
check 3 "The Pod of web runs on cka-sim-worker2" running
check 1 "Deployment web is unchanged and no plugins were copied" untouched
