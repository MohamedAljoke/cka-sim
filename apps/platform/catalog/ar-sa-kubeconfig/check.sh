KC="$COURSE/ci-bot.kubeconfig"
K() { kubectl --kubeconfig "$KC" --request-timeout=5s "$@"; }
whoami() { K auth whoami -o jsonpath='{.status.userInfo.username}' 2>/dev/null; }
is_bot() { eq system:serviceaccount:ar-sa-kubeconfig:ci-bot "$(whoami)"; }
works() { is_bot && K get pods >/dev/null && ! K get secrets >/dev/null 2>&1; }
context_ok() { is_bot && eq "ci-bot ar-sa-kubeconfig" "$(K config view -o jsonpath='{.current-context} {.contexts[?(@.name=="ci-bot")].context.namespace}')" && [ -n "$(K config view --raw -o jsonpath='{.clusters[?(@.name=="cka-sim")].cluster.certificate-authority-data}')" ]; }
lifetime() {
  local payload
  payload=$(K config view --raw -o jsonpath='{.users[0].user.token}' | cut -d. -f2 | tr '_-' '/+')
  while [ $(( ${#payload} % 4 )) -ne 0 ]; do payload="$payload="; done
  echo "$payload" | base64 -d 2>/dev/null | jq '.exp - .iat'
}
one_hour() { is_bot && [ "$(lifetime)" -ge 3000 ] && [ "$(lifetime)" -le 3700 ]; }

check 2 "The kubeconfig logs in as ServiceAccount ci-bot" is_bot
check 1 "It can list Pods in ar-sa-kubeconfig and nothing more" works
check 1 "Context ci-bot is current, defaults to ar-sa-kubeconfig, with the CA embedded" context_ok
check 1 "The token is valid for 1 hour" one_hour
