answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/me.txt" 2>/dev/null | head -1 | tr -d ' \r'; }
me() { kubectl auth whoami -o jsonpath="$1"; }
sorted() { tr ',' '\n' | grep . | sort | paste -sd,; }
user_ok() { eq "$(me '{.status.userInfo.username}')" "$(answer user)"; }
groups_ok() { eq "$(me '{.status.userInfo.groups[*]}' | tr ' ' ',' | sorted)" "$(answer groups | sorted)"; }
listing() { grep -E '^configmaps +\[\] +\[\] +\[get list\]' "$COURSE/viewer.txt" 2>/dev/null; }
viewer_ok() { listing >/dev/null && grep -q 'selfsubjectreviews' "$COURSE/viewer.txt"; }

check 1 "user is the current kubeconfig's user" user_ok
check 1 "groups lists all of its groups" groups_ok
check 2 "viewer.txt is the can-i list of ServiceAccount viewer" viewer_ok
