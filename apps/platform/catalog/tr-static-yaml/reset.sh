for f in /etc/kubernetes/manifests/*; do
  grep -q 'name: node-exporter' "$f" 2>/dev/null && rm -f "$f"
done
exit 0
