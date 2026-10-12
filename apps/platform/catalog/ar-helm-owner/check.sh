answer() { tr -d '[:space:]' < "$COURSE/release.txt" 2>/dev/null; }
listing() { timeout 30 bash "$COURSE/releases.sh" 2>/dev/null; }
all_namespaces() { listing | grep -qw payments && listing | grep -qw audit-trail; }

check 3 "release.txt names payments" eq payments "$(answer)"
check 2 "releases.sh lists the releases of all Namespaces" all_namespaces
