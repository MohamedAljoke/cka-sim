# Sourced before every setup.sh, check.sh, solution.sh and reset.sh. They run as root on the task's host,
# where kubectl is already admin. TASK_ID is set.

# check <points> <description> <command...>
# Prints the PASS/FAIL line the grader reads; the command's own output is thrown away.
check() {
  local points=$1 description=$2
  shift 2
  if "$@" >/dev/null 2>&1; then
    echo "PASS $points $description"
  else
    echo "FAIL $points $description"
  fi
}

eq() { [ "$1" = "$2" ]; }

# wait_for <seconds> <command...>: retry until it succeeds or time runs out.
wait_for() {
  local timeout=$1
  shift
  local deadline=$((SECONDS + timeout))
  until "$@" >/dev/null 2>&1; do
    [ "$SECONDS" -ge "$deadline" ] && return 1
    sleep 2
  done
}

# connects <namespace> <pod> <host> <port>: whether the Pod opens a TCP connection within 2 seconds.
connects() { kubectl -n "$1" exec "$2" -- nc -z -w 2 "$3" "$4"; }

# pod_ip <namespace> <pod>
pod_ip() { kubectl -n "$1" get pod "$2" -o jsonpath='{.status.podIP}'; }

# web_app <namespace> <name>: Deployment and Service <name> on port 80 that answer "<name> <path it got>".
web_app() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: v1
kind: ConfigMap
metadata:
  name: $2-conf
  namespace: $1
data:
  default.conf: |
    server {
      listen 80;
      location / {
        default_type text/plain;
        return 200 "$2 \$request_uri\n";
      }
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $2
  namespace: $1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: $2
  template:
    metadata:
      labels:
        app: $2
    spec:
      containers:
      - name: $2
        image: nginx:1.27-alpine
        ports:
        - containerPort: 80
        volumeMounts:
        - name: conf
          mountPath: /etc/nginx/conf.d
      volumes:
      - name: conf
        configMap:
          name: $2-conf
---
apiVersion: v1
kind: Service
metadata:
  name: $2
  namespace: $1
spec:
  selector:
    app: $2
  ports:
  - port: 80
YAML
  kubectl -n "$1" rollout status deployment "$2" --timeout=120s >/dev/null
}

# gateway <namespace> <name>: a Gateway of class traefik with one listener, "http", on port 80.
gateway() {
  kubectl apply -f - >/dev/null <<YAML
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: $2
  namespace: $1
spec:
  gatewayClassName: traefik
  listeners:
  - name: http
    protocol: HTTP
    port: 80
YAML
}

# condition <namespace> <kind/name> <type>: "<status> <reason>" of a Gateway API condition. A route's
# conditions are those of its first parent.
condition() {
  local path='.status.conditions'
  case $2 in httproute/*) path='.status.parents[0].conditions' ;; esac
  kubectl -n "$1" get "$2" -o jsonpath="{$path[?(@.type==\"$3\")].status} {$path[?(@.type==\"$3\")].reason}"
}

# Traefik serves every Ingress (class nginx) and Gateway (class traefik) in the lab on this address.
TRAEFIK=10.96.80.80

# fetch <host> <path> [curl options...]: the response body of an HTTP request through Traefik.
fetch() { local host=$1 path=$2; shift 2; curl -s -m 5 -H "Host: $host" "$@" "http://$TRAEFIK$path"; }

# fetch_tls <host> <path> [curl options...]: the same over HTTPS, with SNI set to <host>.
fetch_tls() { local host=$1 path=$2; shift 2; curl -sk -m 5 --resolve "$host:443:$TRAEFIK" "$@" "https://$host$path"; }

# served_cert <host>: the subject of the certificate Traefik presents for <host>.
served_cert() { echo | openssl s_client -connect "$TRAEFIK:443" -servername "$1" 2>/dev/null | openssl x509 -noout -subject 2>/dev/null; }

# A pod on a node whose kubelet another task broke never confirms termination and would hold
# the namespace in Terminating forever, so pods are force-deleted first. Their controllers go
# before them, or they would recreate pods while the namespace empties (halves the wait).
# The label lets the server delete the namespace once the task is done (tasks.TidyScript).
# Pods are refused until the controller manager has made the namespace's default ServiceAccount.
fresh_ns() {
  kubectl -n "$1" delete deployments,replicasets,statefulsets,daemonsets,jobs,cronjobs --all --wait=false >/dev/null 2>&1
  kubectl -n "$1" delete pods --all --force --grace-period=0 >/dev/null 2>&1
  kubectl delete namespace "$1" --ignore-not-found --wait=true --timeout=120s >/dev/null
  kubectl create namespace "$1" >/dev/null
  kubectl label namespace "$1" cka-sim/task="$TASK_ID" >/dev/null
  wait_for 60 kubectl -n "$1" get serviceaccount default
}

# Exam-style answer files live in /opt/course/<task id>/ on the host.
COURSE="/opt/course/$TASK_ID"

# The candidate writes there without sudo, as in the exam.
fresh_course() {
  rm -rf "$COURSE"
  mkdir -p "$COURSE"
  chmod 0777 "$COURSE"
}

# The exam's control plane has etcdctl and etcdutl; kind's node image doesn't, so copy them out
# of the etcd image the cluster already runs.
etcd_tools() {
  command -v etcdutl >/dev/null && return
  local image mnt
  image=$(ctr -n k8s.io images ls -q | grep -m1 '^registry.k8s.io/etcd:')
  mnt=$(mktemp -d)
  ctr -n k8s.io images mount "$image" "$mnt" >/dev/null
  cp "$mnt/usr/local/bin/etcdctl" "$mnt/usr/local/bin/etcdutl" /usr/local/bin/
  ctr -n k8s.io images unmount "$mnt" >/dev/null
  rmdir "$mnt"
}
