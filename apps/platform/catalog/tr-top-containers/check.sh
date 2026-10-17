answer() { tr -d '[:space:]' < "$COURSE/container.txt" 2>/dev/null; }

check 3 "container.txt names cache" eq cache "$(answer)"
