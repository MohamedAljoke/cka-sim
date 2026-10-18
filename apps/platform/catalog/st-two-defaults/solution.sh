kubectl annotate storageclass legacy storageclass.kubernetes.io/is-default-class=false --overwrite
kubectl get storageclass
printf 'omitted: standard\nempty: none\n' > /opt/course/st-two-defaults/classes.txt
