answer() { sed -n "s/^$1: *\(.*\)$/\1/p" "$COURSE/answers.txt" 2>/dev/null | head -1 | tr -d ' \r' | tr 'A-Z' 'a-z'; }
is() { eq "$2" "$(answer "$1")"; }
same_ns() { is 1 allowed && is 2 denied && is 3 allowed; }
reverse() { is 4 denied; }
other_ns() { is 5 denied; }
port() { is 6 denied; }

check 2 "1 allowed, 2 denied, 3 allowed" same_ns
check 1 "4 denied: web has no rule that admits api" reverse
check 1 "5 denied: a podSelector alone only matches the same namespace" other_ns
check 1 "6 denied: only port 5432 is open on db" port
