for f in /etc/kubernetes/manifests/*; do
  grep -q 'name: static-web' "$f" 2>/dev/null && rm -f "$f"
done
fresh_ns wl-nodename
