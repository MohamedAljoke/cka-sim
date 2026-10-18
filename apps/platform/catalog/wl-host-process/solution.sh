id=$(crictl ps -q --name '^proc$' --namespace wl-host-process | head -1)
pid=$(crictl inspect "$id" | jq -r .info.pid)
ps -o pid,uid,args -p "$pid"
{
  echo "pid: $pid"
  echo "uid: $(ps -o uid= -p "$pid" | xargs)"
  echo "command: $(ps -o args= -p "$pid")"
} > /opt/course/wl-host-process/process.txt
