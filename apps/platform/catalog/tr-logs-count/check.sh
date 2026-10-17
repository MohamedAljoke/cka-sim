NS=tr-logs-count
expected() {
  local p
  for p in $(kubectl -n $NS get pods -l app=worker -o jsonpath='{.items[*].metadata.name}'); do
    echo "$p: $(kubectl -n $NS logs "$p" | grep -c ERROR)"
  done | sort
}
written() { tr -d '\r' < "$COURSE/errors.txt" 2>/dev/null | sed 's/: */: /' | grep . | sort; }

check 3 "errors.txt has the right ERROR count for every Pod of worker" eq "$(expected)" "$(written)"
