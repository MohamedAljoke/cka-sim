answer() { tr -d '[:space:]' < "$COURSE/$1" 2>/dev/null; }

check 2 "cpu.txt names boreas" eq boreas "$(answer cpu.txt)"
check 2 "memory.txt names atlas" eq atlas "$(answer memory.txt)"
