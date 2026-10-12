FILE="$COURSE/events.txt"
first() { grep -n "$1" "$FILE" | head -1 | cut -d: -f1; }
last() { grep -n "$1" "$FILE" | tail -1 | cut -d: -f1; }
both() { grep -q Scheduled "$FILE" && grep -q FailedMount "$FILE" && grep -q report "$FILE"; }
no_ledger() { ! grep -q ledger "$FILE"; }
in_order() { [ "$(first Scheduled)" -lt "$(last FailedMount)" ] 2>/dev/null; }

check 2 "events.txt has the Scheduled and FailedMount events of report" both
if both; then
  check 1 "events.txt has no events of Pod ledger" no_ledger
  check 1 "Scheduled comes before FailedMount" in_order
else
  echo "FAIL 1 events.txt has no events of Pod ledger"
  echo "FAIL 1 Scheduled comes before FailedMount"
fi
