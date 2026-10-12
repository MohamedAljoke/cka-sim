expected() {
  kubectl get pods -A -o jsonpath='{range .items[?(@.status.qosClass=="BestEffort")]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}' |
    grep -E '^tr-besteffort-(a|b)/' | sort | xargs
}
written() { tr -d ' \r' < "$COURSE/pods.txt" 2>/dev/null | grep . | sort | xargs; }

check 3 "pods.txt lists exactly the BestEffort Pods of both namespaces" eq "$(expected)" "$(written)"
