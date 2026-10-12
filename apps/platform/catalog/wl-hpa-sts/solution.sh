kubectl -n wl-hpa-sts autoscale statefulset db --min=1 --max=3 --cpu=80% >/dev/null
echo 3 > /opt/course/wl-hpa-sts/pvcs.txt
echo db-2 > /opt/course/wl-hpa-sts/first-removed.txt
