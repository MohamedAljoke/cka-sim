answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/kubelet.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
config_file() { systemctl cat kubelet | grep -o -- '--config=[^ "]*' | tail -1 | cut -d= -f2; }
setting() { sed -n "s/^$1: *//p" "$(config_file)" | tr -d ' \r'; }
config_ok() { eq "$(config_file)" "$(answer config)"; }
static_ok() { eq "$(setting staticPodPath)" "$(answer staticPodPath | sed 's|/$||')"; }
cgroup_ok() { eq "$(setting cgroupDriver)" "$(answer cgroupDriver)"; }

check 1 "config is the kubelet's config file" config_ok
check 1 "staticPodPath is right" static_ok
check 1 "cgroupDriver is right" cgroup_ok
