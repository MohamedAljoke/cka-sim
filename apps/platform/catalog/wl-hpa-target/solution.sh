kubectl -n wl-hpa-target patch hpa orders --type merge -p '{"spec":{"scaleTargetRef":{"kind":"Deployment"}}}' >/dev/null
