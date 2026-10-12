for ns in tr-besteffort-a tr-besteffort-b; do
  kubectl -n $ns get pods -o custom-columns=NAME:.metadata.name,QOS:.status.qosClass --no-headers |
    awk -v ns=$ns '$2=="BestEffort"{print ns"/"$1}'
done > /opt/course/tr-besteffort/pods.txt
