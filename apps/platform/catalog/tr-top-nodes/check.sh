listing() { timeout 30 bash "$COURSE/nodes.sh" 2>/dev/null | grep '^cka-sim-'; }
every_node() { eq 3 "$(listing | wc -l)"; }
by_memory() { every_node && listing | awk '{print $4 + 0}' | sort -c -n -r; }

check 1 "nodes.sh lists every node" every_node
check 2 "nodes.sh sorts them by memory usage, highest first" by_memory
