FILE="$COURSE/app-errors.log"
expected() { kubectl -n tr-container-logs logs checkout -c app | grep ERROR; }
matches() { eq "$(expected)" "$(cat "$FILE")"; }
no_proxy() { ! grep -q upstream "$FILE"; }

check 3 "app-errors.log holds exactly the ERROR lines of container app" matches
if [ -s "$FILE" ]; then
  check 1 "No line from container proxy" no_proxy
else
  echo "FAIL 1 No line from container proxy"
fi
