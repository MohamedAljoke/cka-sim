kubectl -n wl-hpa autoscale deployment api --min=2 --max=6 --cpu=60% >/dev/null
