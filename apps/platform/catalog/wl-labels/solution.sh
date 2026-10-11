kubectl -n wl-labels get pods -l tier=backend,env=prod -o name | cut -d/ -f2 > /opt/course/wl-labels/backend-prod.txt
kubectl -n wl-labels label pods -l app=web team=payments
kubectl -n wl-labels label pod web-dev env-
kubectl -n wl-labels annotate pod api owner=platform
