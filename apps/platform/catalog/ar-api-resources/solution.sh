# kubectl api-resources | grep -E '^(deployments|leases|...)\s' -> APIVERSION and NAMESPACED columns
for r in deployments leases priorityclasses storageclasses events csinodes; do
  kubectl api-resources --no-headers | awk -v r="$r" '$1==r { gv=(NF==5)?$3:$2; ns=(NF==5)?$4:$3; print r": "gv" "(ns=="true"?"namespaced":"cluster"); exit }'
done > /opt/course/ar-api-resources/resources.txt
