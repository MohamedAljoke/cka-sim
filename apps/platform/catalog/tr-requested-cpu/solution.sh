# kubectl describe nodes | grep -A8 'Allocated resources' -> compare the cpu Requests column per node
for n in $(kubectl get nodes -o name); do
  echo "$(kubectl describe "$n" | awk '/Allocated resources/{f=1} f && $1=="cpu"{print $2; exit}' | tr -d m) ${n#node/}"
done | sort -n | tail -1 | awk '{print $2}' > /opt/course/tr-requested-cpu/node.txt
