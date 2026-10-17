answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/modes.txt" 2>/dev/null | head -1 | tr -d ' \r'; }

check 1 "1 is ReadWriteOncePod" eq ReadWriteOncePod "$(answer 1)"
check 1 "2 is ReadOnlyMany" eq ReadOnlyMany "$(answer 2)"
check 1 "3 is ReadWriteMany" eq ReadWriteMany "$(answer 3)"
check 1 "4 is ReadWriteOnce" eq ReadWriteOnce "$(answer 4)"
