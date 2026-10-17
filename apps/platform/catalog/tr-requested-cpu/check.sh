busiest() {
  kubectl get pods -A --field-selector status.phase=Running -o json |
    jq -r '[.items[] | {node: .spec.nodeName, cpu: ([.spec.containers[].resources.requests.cpu // "0" | if endswith("m") then (.[:-1] | tonumber) else (tonumber * 1000) end] | add)}]
      | group_by(.node) | map({node: .[0].node, cpu: (map(.cpu) | add)}) | max_by(.cpu) | .node'
}
answer() { tr -d ' \r\n' < "$COURSE/node.txt" 2>/dev/null; }

check 3 "node.txt names the node with the most CPU requested" eq "$(busiest)" "$(answer)"
