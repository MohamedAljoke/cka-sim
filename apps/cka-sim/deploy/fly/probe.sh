#!/usr/bin/env bash
# Tries, from inside one exam VM, everything a user there should not reach: the internet and the other
# VMs' dockerd. A claimed VM should show every line blocked except dns (Fly's resolver can't be locked).
#
#   probe.sh MACHINE_ID
set -euo pipefail

APP=${FLY_LABS_APP:-cka-sim-labs-maljoke}
ID=${1:?which machine? see pool.sh list}
TOKEN=$(fly auth token 2>/dev/null | tail -1)

checks='t() { if timeout 6 sh -c "$2" >/dev/null 2>&1; then echo "$1 OPEN"; else echo "$1 blocked"; fi; }
t internet-ipv4:80 "nc -w 4 1.1.1.1 80 </dev/null"
t internet-ipv4:443 "nc -w 4 1.1.1.1 443 </dev/null"
t internet-ipv6:443 "nc -w 4 2606:4700:4700::1111 443 </dev/null"
t dns "nslookup example.com fdaa::3"'
while read -r other ip; do
  [ "$other" = "$ID" ] && continue
  checks+=$'\n'"t vm-$other:2375 \"nc -w 4 $ip 2375 </dev/null\""
done < <(fly machine list -a "$APP" --json | jq -r '.[] | select(.state == "started") | [.id, .private_ip] | @tsv')

curl -fsS -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  "https://api.machines.dev/v1/apps/$APP/machines/$ID/exec" \
  -d "$(jq -n --arg c "$checks" '{command: ["sh", "-c", $c], timeout: 90}')" | jq -r '.stdout'
