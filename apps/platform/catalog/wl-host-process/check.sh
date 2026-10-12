answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/process.txt" 2>/dev/null | head -1 | tr -d '\r'; }
pid() { crictl inspect "$(crictl ps -q --name '^proc$' --namespace wl-host-process | head -1)" | jq -r '.info.pid'; }
uid() { awk '/^Uid:/{print $2}' "/proc/$(pid)/status"; }

check 2 "pid is the container's main process on the host" eq "$(pid)" "$(answer pid | xargs)"
check 1 "uid is 3000" eq "$(uid)" "$(answer uid | xargs)"
check 1 "command is sleep 86401" eq "sleep 86401" "$(answer command | xargs)"
